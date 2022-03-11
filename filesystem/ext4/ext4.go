package ext4

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/diskfs/go-diskfs/filesystem"
	"github.com/diskfs/go-diskfs/util"
	"github.com/lunixbochs/struc"
)

type FileSystem struct {
	sb  *Superblock
	dev *os.File
	start int64
}

func Read(file util.File, size int64, start int64, blocksize int64) (*FileSystem, error) {
	var (
		err  error
	)

	_, err = file.Seek(start+Superblock0Offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	sb := &Superblock{}
	err = struc.Unpack(file, sb)
	if err != nil {
		return nil, err
	}

	fs := &FileSystem{
		sb: sb,
		dev: file.(*os.File),
		start: start,
	}

	return fs, nil
}

func (fs *FileSystem) Superblock() *Superblock {
	return fs.sb
}

func (fs *FileSystem) Label() string {
	return string(fs.sb.Volume_name[:])
}

// Mkdir make a directory at the given path. It is equivalent to `mkdir -p`, i.e. idempotent, in that:
//
// * It will make the entire tree path if it does not exist
// * It will not return an error if the path already exists
func (fs *FileSystem) Mkdir(path string) error {
	return fs.mkDir(path, 0775)
}

// OpenFile returns an io.ReadWriter from which you can read the contents of a file
// or write contents to the file
//
// accepts os.OpenFile flags: O_CREATE, O_APPEND, and always O_RDWR
//
// returns an error if the file does not exist
func (fs *FileSystem) OpenFile(p string, flag int) (filesystem.File, error) {
	// get the path
	dir := path.Dir(p)
	filename := path.Base(p)
	// if the dir == filename, then it is just /
	if dir == filename {
		return nil, fmt.Errorf("Cannot open directory %s as file", p)
	}

	inodeNum := int64(ROOT_INO)
	var inode *Inode
	parts := strings.Split(dir, "/")
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		inode = fs.getInode(inodeNum)
		dirContents := inode.ReadDirectory()
		found := false
		for i := 0; i < len(dirContents); i++ {
			//log.Println(string(dirContents[i].Name), part, dirContents[i].Flags, dirContents[i].Inode)
			if string(dirContents[i].Name) == part {
				found = true
				inodeNum = int64(dirContents[i].Inode)
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("No such directory %s", dir)
		}
	}

	fileExists := false
	inode = fs.getInode(inodeNum)
	lastDirContents := inode.ReadDirectory()
	for i := 0; i < len(lastDirContents); i++ {
		if string(lastDirContents[i].Name) == filename {
			fileExists = true
			inodeNum = int64(lastDirContents[i].Inode)
			break
		}
	}

	if fileExists {
		inode = fs.getInode(inodeNum)

		pos := int64(0)
		if flag&os.O_APPEND == os.O_APPEND {
			pos = inode.GetSize()
		}

		//log.Printf("Inode %d with mode %x", inode.num, inode.Mode)
		return &File{extFile{
			fs:    fs,
			inode: inode,
			pos:   pos,
		}}, nil
	}

	// create when not exists
	if flag&os.O_CREATE == 0 {
		return nil, fmt.Errorf("Target file %s does not exist and was not asked to create", p)
	}

	newFile := fs.CreateNewFile(0777)
	log.Printf("Creating new file with inode %d and perms %d", newFile.inode.num, newFile.inode.Mode)
	newFile.inode.Mode |= 0x8000
	newFile.inode.UpdateCsumAndWriteback()

	NewDirectory(inode).AddEntry(&DirectoryEntry2{
		Inode: uint32(newFile.inode.num),
		Flags: 0,
		Name:  filename,
	})

	return newFile, nil
}

// ReadDir return the contents of a given directory in a given filesystem.
//
// Returns a slice of os.FileInfo with all of the entries in the directory.
//
// Will return an error if the directory does not exist or is a regular file and not a directory
func (fs *FileSystem) ReadDir(dir string) ([]os.FileInfo, error) {
	
	inodeNum := int64(ROOT_INO)
	var inode *Inode
	parts := strings.Split(dir, "/")
	//if len(parts) == 0

	for _, part := range parts[1:] {
		if len(part) == 0 {
			continue
		}

		inode = fs.getInode(inodeNum)
		dirContents := inode.ReadDirectory()
		found := false
		for i := 0; i < len(dirContents); i++ {
			//log.Println(string(dirContents[i].Name), part, dirContents[i].Flags, dirContents[i].Inode)
			if string(dirContents[i].Name) == part {
				found = true
				inodeNum = int64(dirContents[i].Inode)
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("No such directory %s", dir)
		}
	}

	inode = fs.getInode(inodeNum)
	lastDirContents := inode.ReadDirectory()
	count := len(lastDirContents)
	ret := make([]os.FileInfo, count, count)
	for i := 0; i < len(lastDirContents); i++ {
		inodeNum = int64(lastDirContents[i].Inode)
		inode = fs.getInode(inodeNum)

		isDir := false
		if inode.Mode&0x4000 == 0x4000 {
			isDir = true
		}

		ret[i] = FileInfo{
			modTime: time.Unix(int64(inode.Mtime), 0),
			//mode: os.FileMode(inode.Mode),
			name:  lastDirContents[i].Name,
			size:  inode.GetSize(),
			isDir: isDir,
		}
	}

	return ret, nil
}

// Type returns the type code for the filesystem. Always returns filesystem.TypeFat32
func (fs *FileSystem) Type() filesystem.Type {
	return filesystem.TypeExt4
}

func (fs *FileSystem) create(path string) (*File, error) {
	log.Println("CREATE", path)
	parts := strings.Split(path, "/")

	inode := fs.getInode(int64(ROOT_INO))

	for _, part := range parts[:len(parts)-1] {
		if len(part) == 0 {
			continue
		}

		dirContents := inode.ReadDirectory()
		found := false
		for i := 0; i < len(dirContents); i++ {
			//log.Println(string(dirContents[i].Name), part, dirContents[i].Flags, dirContents[i].Inode)
			if string(dirContents[i].Name) == part {
				found = true
				inode = fs.getInode(int64(dirContents[i].Inode))
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("No such file or directory")
		}
	}

	name := parts[len(parts)-1]

	newFile := fs.CreateNewFile(0777)
	log.Printf("Creating new file with inode %d and perms %d", newFile.inode.num, newFile.inode.Mode)
	newFile.inode.Mode |= 0x8000
	newFile.inode.UpdateCsumAndWriteback()

	NewDirectory(inode).AddEntry(&DirectoryEntry2{
		Inode: uint32(newFile.inode.num),
		Flags: 0,
		Name:  name,
	})

	return newFile, nil
}

func (fs *FileSystem) open(name string) (*File, error) {
	parts := strings.Split(name, "/")

	inodeNum := int64(ROOT_INO)
	var inode *Inode
	for _, part := range parts {
		if len(part) == 0 {
			continue
		}

		inode = fs.getInode(inodeNum)
		dirContents := inode.ReadDirectory()
		found := false
		for i := 0; i < len(dirContents); i++ {
			//log.Println(string(dirContents[i].Name), part, dirContents[i].Flags, dirContents[i].Inode)
			if string(dirContents[i].Name) == part {
				found = true
				inodeNum = int64(dirContents[i].Inode)
				break
			}
		}

		if !found {
			return nil, fmt.Errorf("No such file or directory")
		}
	}

	inode = fs.getInode(inodeNum)
	//log.Printf("Inode %d with mode %x", inode.num, inode.Mode)
	return &File{extFile{
		fs:    fs,
		inode: inode,
		pos:   0,
	}}, nil
}

func (fs *FileSystem) Remove(name string) error {
	return nil
}

func (fs *FileSystem) mkDir(path string, perm os.FileMode) error {
	log.Println("MKDIR", path)
	parts := strings.Split(path, "/")

	inode := fs.getInode(int64(ROOT_INO))

	for _, part := range parts[:len(parts)-1] {
		if len(part) == 0 {
			continue
		}

		dirContents := inode.ReadDirectory()
		found := false
		for i := 0; i < len(dirContents); i++ {
			//log.Println(string(dirContents[i].Name), part, dirContents[i].Flags, dirContents[i].Inode)
			if string(dirContents[i].Name) == part {
				found = true
				inode = fs.getInode(int64(dirContents[i].Inode))
				break
			}
		}

		if !found {
			return fmt.Errorf("No such file or directory")
		}
	}

	name := parts[len(parts)-1]

	newFile := fs.CreateNewFile(perm)
	log.Printf("Creating new directory with inode %d and perms %d", newFile.inode.num, newFile.inode.Mode)
	newFile.inode.Mode |= 0x4000
	newFile.inode.UpdateCsumAndWriteback()

	{
		checksummer := NewChecksummer(inode.fs.sb)
		checksummer.Write(inode.fs.sb.Uuid[:])
		checksummer.WriteUint32(uint32(newFile.inode.num))
		checksummer.WriteUint32(uint32(newFile.inode.Generation))

		dirEntryDot := DirectoryEntry2{
			Inode:   uint32(newFile.inode.num),
			Flags:   2,
			Rec_len: 12,
			Name:    ".",
		}
		recLenDot, _ := struc.Sizeof(&dirEntryDot)
		struc.Pack(checksummer, dirEntryDot)
		struc.Pack(newFile, dirEntryDot)
		{
			blank1 := make([]byte, 12-recLenDot)
			checksummer.Write(blank1)
			newFile.Write(blank1)
		}

		dirEntryDotDot := DirectoryEntry2{
			Inode: uint32(inode.num),
			Flags: 2,
			Name:  "..",
		}
		recLenDotDot, _ := struc.Sizeof(&dirEntryDotDot)
		dirEntryDotDot.Rec_len = uint16(1024 - 12 - 12)
		struc.Pack(checksummer, dirEntryDotDot)
		struc.Pack(newFile, dirEntryDotDot)

		blank := make([]byte, 1024-12-12-recLenDotDot)
		checksummer.Write(blank)
		newFile.Write(blank)

		dirSum := DirectoryEntryCsum{
			FakeInodeZero: 0,
			Rec_len:       uint16(12),
			FakeName_len:  0,
			FakeFileType:  0xDE,
			Checksum:      checksummer.Get(),
		}
		struc.Pack(newFile, &dirSum)
	}

	NewDirectory(inode).AddEntry(&DirectoryEntry2{
		Inode: uint32(newFile.inode.num),
		Flags: 0,
		Name:  name,
	})

	newFile.inode.Links_count++
	newFile.inode.UpdateCsumAndWriteback()

	inode.Links_count++
	inode.UpdateCsumAndWriteback()

	bgd := fs.getBlockGroupDescriptor((newFile.inode.num - 1) / int64(inode.fs.sb.InodePer_group))
	bgd.Used_dirs_count_lo++
	bgd.UpdateCsumAndWriteback()

	return nil
}

func (fs *FileSystem) Close() error {
	err := fs.dev.Close()
	if err != nil {
		return err
	}
	fs.sb = nil
	fs.dev = nil
	return nil
}

// --------------------------
func (fs *FileSystem) GetInode(inodeAddress int64) *Inode {
	return fs.getInode(inodeAddress)
}

func (fs *FileSystem) getInode(inodeAddress int64) *Inode {
	bgd := fs.getBlockGroupDescriptor((inodeAddress - 1) / int64(fs.sb.InodePer_group))
	index := (inodeAddress - 1) % int64(fs.sb.InodePer_group)
	pos := bgd.GetInodeTableLoc()*fs.sb.GetBlockSize() + index*int64(fs.sb.Inode_size)
	//log.Printf("%d %d %d %d", bgd.GetInodeTableLoc(), fs.sb.GetBlockSize(), index, fs.sb.Inode_size)
	fs.dev.Seek(fs.start + pos, io.SeekStart)

	inode := &Inode{
		fs:      fs,
		address: pos,
		num:     inodeAddress}
	struc.Unpack(fs.dev, &inode)
	//log.Printf("Read inode %d, contents:\n%+v\n", inodeAddress, inode)
	return inode
}

func (fs *FileSystem) getBlockGroupDescriptor(blockGroupNum int64) *GroupDescriptor {
	blockSize := fs.sb.GetBlockSize()
	bgdtLocation := 1024/blockSize + 1

	size := int64(32)
	if fs.sb.FeatureIncompat64bit() {
		size = int64(64)
	}
	addr := bgdtLocation*blockSize + size*blockGroupNum
	bgd := &GroupDescriptor{
		fs:      fs,
		address: addr,
		num:     blockGroupNum,
	}
	fs.dev.Seek(fs.start + addr, 0)
	struc.Unpack(io.LimitReader(fs.dev, size), &bgd)
	//log.Printf("Read block group %d, contents:\n%+v\n", blockGroupNum, bgd)
	return bgd
}

func (fs *FileSystem) CreateNewFile(perm os.FileMode) *File {
	var inode *Inode
	for i := int64(0); i < fs.sb.numBlockGroups; i++ {
		bgd := fs.getBlockGroupDescriptor(i)
		inode = bgd.GetFreeInode()
		if inode != nil {
			break
		}
	}

	if inode == nil {
		log.Fatalln("Couldn't get free inode", fs.sb.numBlockGroups, fs.sb.Free_inodeCount)
		return nil
	}

	inode.Mode = uint16(perm & 0x1FF)
	inode.UpdateCsumAndWriteback()

	return &File{extFile{
		fs:    fs,
		inode: inode,
	}}
}

func (fs *FileSystem) GetFreeBlocks(n int) (int64, int64) {
	for i := int64(0); i < fs.sb.numBlockGroups; i++ {
		bgd := fs.getBlockGroupDescriptor(i)
		blockNum, numBlocks := bgd.GetFreeBlocks(int64(n))
		if blockNum > 0 {
			return blockNum + i*int64(fs.sb.BlockPer_group), numBlocks
		}
	}
	log.Fatalf("Failed to find free block")
	return 0, 0
}
