package pkg

import (
	"os"
	"regexp"

	godisk "github.com/diskfs/go-diskfs"
	"github.com/diskfs/go-diskfs/filesystem"
	yaml "gopkg.in/yaml.v2"
)

type ReplaceItem struct {
	MatchRegexp string `yaml:"matchRegexp"`
	ReplaceStr  string `yaml:"replaceStr"`
}

type ModiFile struct {
	FileName     string        `yaml:"fileName"`
	ReplaceItems []*ReplaceItem `yaml:"replaceItem"`
}

type ModiTask struct {
	ModiFiles    []*ModiFile `yaml:"modiFile"`
	Image        string     `yaml:"-"`
	PartitionNum int        `yaml:"-"`
}

func CreateModiTaskFromYaml(yfile string) (*ModiTask, error) {
	bs, err := os.ReadFile(yfile)
	if err != nil {
		return nil, err
	}

	t := &ModiTask{}

	if err = yaml.Unmarshal(bs, t); err != nil {
		return nil, err
	}

	return t, nil
}

func (t *ModiTask) SaveToYaml(yfile string) error {
	bs, err := yaml.Marshal(t)
	if err != nil {
		return err
	}

	return os.WriteFile(yfile, bs, 0640)
}

func (t *ModiTask) Modify(fs filesystem.FileSystem) (error) {
	for _, fi := range t.ModiFiles {
		modiFi, s, err := ReadFile3(fs, fi.FileName)
		if err != nil {
			return err
		}
		defer modiFi.Close()

		for _, itm := range fi.ReplaceItems {
			s, err = replace(s, itm.MatchRegexp, itm.ReplaceStr)
			if err != nil {
				return err
			}
		}

		if _, err := modiFi.Seek(0, 0); err != nil {
			return err
		}

		if _, err := modiFi.Write([]byte(s)); err != nil {
			return err
		}
	}

	return nil
}

func replace(origStr, regexStr, replaceStr string) (string, error) {
	re, err := regexp.Compile(regexStr)
	if err != nil {
		return "", err
	}

	s := re.ReplaceAllString(origStr, replaceStr)
	return s, nil
}

func ReadFile(imageName string, pNum int, fileName string) (string, error) {
	disk, err := godisk.Open(imageName)
	if err != nil {
		return "", err
	}
	defer disk.File.Close()

	fs, err := disk.GetFilesystem(pNum)
	if err != nil {
		return "", err
	}

	return ReadFile2(fs, fileName)
}

func ReadFile2(fs filesystem.FileSystem, fileName string) (string, error) {
	f1, err := fs.OpenFile(fileName, os.O_RDONLY)
	if err != nil {
		return "", err
	}
	defer f1.Close()

	f1Content := make([]byte, 64*1<<10)

	n, err := f1.Read(f1Content)
	if err != nil {
		return "", err
	}

	return string(f1Content[:n]), nil
}

func ReadFile3(fs filesystem.FileSystem, fileName string) (filesystem.File, string, error) {
	f1, err := fs.OpenFile(fileName, os.O_RDWR)
	if err != nil {
		return nil, "", err
	}

	f1Content := make([]byte, 64*1<<10)

	n, err := f1.Read(f1Content)
	if err != nil {
		return nil, "", err
	}

	return f1, string(f1Content[:n]), nil
}