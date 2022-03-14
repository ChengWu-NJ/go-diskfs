package diskfs

import (
	"os"
	"testing"

	"github.com/diskfs/go-diskfs/filesystem/ext4"
)

func TestGetBlockGroupDescriptor(t *testing.T) {
	disk, err := Open(`/tmp/dietpi.img`)
	if err != nil {
		t.Error(err)
	}
	defer disk.File.Close()

	fs, err := disk.GetFilesystem(1)
	if err != nil {
		t.Error(err)
	}
	t.Log(fs)

	bgd := fs.(*ext4.FileSystem).GetBlockGroupDescriptor(0)
	t.Logf("\n%#v\n", bgd)
}

func TestGetInode(t *testing.T) {
	disk, err := Open(`/tmp/dietpi.img`)
	if err != nil {
		t.Error(err)
	}
	defer disk.File.Close()

	fs, err := disk.GetFilesystem(1)
	if err != nil {
		t.Error(err)
	}
	t.Log(fs)

	inode := fs.(*ext4.FileSystem).GetInode(2) // root /
	t.Logf("inode 2:[%#v]\n", inode)

	dirContents := inode.ReadDirectory()
	for i, e := range dirContents {
		t.Logf("%d: [%#v]\n", i, e)
	}
}

func TestCreateFile(t *testing.T) {
	disk, err := Open(`/tmp/dietpi.img`)
	if err != nil {
		t.Error(err)
	}
	defer disk.File.Close()

	fs, err := disk.GetFilesystem(1)
	if err != nil {
		t.Error(err)
	}

	f1Name := `/etc/c3p1.conf`
	f1, err := fs.OpenFile(f1Name, os.O_CREATE|os.O_RDWR)
	if err != nil {
		t.Error(err)
	}
	defer f1.Close()

	f1Content := `userid = "ABCD"
	
validPeriodOfDays = 100
`
	n, err := f1.Write([]byte(f1Content))
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s was written with length %d", f1Name, n)
}

func TestReadFile(t *testing.T) {
	disk, err := Open(`/tmp/dietpi.img`)
	if err != nil {
		t.Error(err)
	}
	defer disk.File.Close()

	fs, err := disk.GetFilesystem(1)
	if err != nil {
		t.Error(err)
	}

	f1Name := `/etc/c3p1.conf`
	f1, err := fs.OpenFile(f1Name, os.O_RDONLY)
	if err != nil {
		t.Error(err)
	}
	defer f1.Close()

	f1Content := make([]byte, 64*1<<10)

	n, err := f1.Read(f1Content)
	if err != nil {
		t.Error(err)
	}

	t.Logf("%s was read with length:%d and content:\n[\n%s\n]\n", f1Name, n, string(f1Content))
}

func TestModifyFile(t *testing.T) {

	disk, err := Open(`/tmp/dietpi.img`)
	if err != nil {
		t.Error(err)
	}
	defer disk.File.Close()
	//t.Logf("disk:[%#v]\n", disk)

	/*
	tab, err := disk.GetPartitionTable()
	if err != nil {
		t.Error(err)
	}

		pttions := tab.GetPartitions()
		if len(pttions) == 0 {
			t.Error("the image file has no partition\n")
		}
		for i, pt := range pttions {
			t.Logf("P%d:[%#v]\n", i+1, pt)
		}
	*/

	fs, err := disk.GetFilesystem(1)
	if err != nil {
		t.Error(err)
	}

	/*
		sb := fs.(*ext4.FileSystem).Superblock()
		t.Logf("Magic:[%#v]\n", sb.Magic)
		t.Logf("superblock:[%#v]\n", sb)

		files, err := fs.ReadDir("/boot")
		if err != nil {
			t.Error(err)
		}
	*/

	f1Name := `/boot/dietpi.txt`
	f1, err := fs.OpenFile(f1Name, os.O_RDWR)
	if err != nil {
		t.Error(err)
	}
	defer f1.Close()

	buf := make([]byte, 10) //only first 10
	n, err := f1.Read(buf)
	if err != nil {
		t.Error(err)
	}
	t.Logf("%s -- first %d character:[%s]\n", f1Name, n, string(buf))

	buf[1] = (buf[1] + 1) % 127

	_, err = f1.Seek(0, 0)
	if err != nil {
		t.Error(err)
	}

	n, err = f1.Write(buf[:n])
	if err != nil {
		t.Error(err)
	}
	
	_, err = f1.Seek(0, 0)
	if err != nil {
		t.Error(err)
	}

	n, err = f1.Read(buf)
	if err != nil {
		t.Error(err)
	}
	t.Logf("after modify, %s -- first %d character:[%s]\n", f1Name, n, string(buf))
}
