package diskfs

import (
	"testing"

	"github.com/diskfs/go-diskfs/filesystem/ext4"
)

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

	inode := fs.(*ext4.FileSystem).GetInode(10929) // root /
	t.Logf("inode 2:[%#v]\n", inode)

	dirContents := inode.ReadDirectory()
	for i, e := range dirContents {
		t.Logf("%d: [%#v]\n", i, e)
	}
}

func TestExt4(t *testing.T) {
	disk, err := Open(`/tmp/dietpi.img`)
	if err != nil {
		t.Error(err)
	}
	defer disk.File.Close()
	t.Logf("disk:[%#v]\n", disk)

	tab, err := disk.GetPartitionTable()
	if err != nil {
		t.Error(err)
	}

	pttions := tab.GetPartitions()
	if len(pttions) == 0 {
		t.Error("the image file has no partition\n")
	}
	//for i, pt := range pttions {
	//	t.Logf("P%d:[%#v]\n", i+1, pt)
	//}

	fs, err := disk.GetFilesystem(1)
	if err != nil {
		t.Error(err)
	}

	sb := fs.(*ext4.FileSystem).Superblock()
	t.Logf("Magic:[%#v]\n", sb.Magic)
	//t.Logf("superblock:[%#v]\n", sb)

	files, err := fs.ReadDir("/boot")
	if err != nil {
		t.Error(err)
	}

	t.Log(files)

}
