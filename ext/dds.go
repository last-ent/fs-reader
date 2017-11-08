package ext

import (
	"bytes"
	"fmt"
	"os"
	"unsafe"
)

type BlockGroup struct {
	SuperBlock       *Ext2Sb
	GroupDescriptors *Ext2Gd
	DataBlockBitmap  []byte
	InodeBitmap      []byte
	InodeTable       []*Ext2Inode
}

func INodesPerBlock(blockSize int) int {
	return blockSize / 128
}

func LoadBlockGroup(file *os.File) *BlockGroup {
	blockGroup := &BlockGroup{
		// ignore bootblock 0...1023
		SuperBlock:       LoadSBStruct(file),    // 1024...2047
		GroupDescriptors: LoadGDStruct(file),    // 2048...3071
		DataBlockBitmap:  Load1KB(file, 1024*3), // 3072...4095
		InodeBitmap:      Load1KB(file, 1024*4), // 4096...5120
	}
	INodesCount := int(blockGroup.SuperBlock.SInodesPerGroup) / INodesPerBlock(1024)

	blockGroup.InodeTable = LoadINDTable(
		file,
		blockGroup,
		INodesCount,
	)
	return blockGroup
}

func Load1KB(file *os.File, from int64) []byte {
	return LoadNBytes(file, from, 1024)
}

func LoadNBytes(file *os.File, from int64, size int) []byte {
	block := make([]byte, size)
	file.ReadAt(block, from)
	return block
}

type Ext2Sb struct {
	// Taken from https://github.com/ctdk/sbinfo/blob/master/ext2.go
	SInodesCount          uint32
	SBlocksCount          uint32
	SRBlocksCount         uint32
	SFreeBlocksCount      uint32
	SFreeInodesCount      uint32
	SFirstDataBlock       uint32
	SLogBlockSize         uint32
	SLogClusterSize       uint32
	SBlocksPerGroup       uint32
	SClustersPerGroup     uint32
	SInodesPerGroup       uint32
	SMtime                uint32
	SWtime                uint32
	SMntCount             uint16
	SMaxMntCount          uint16
	SMagic                uint16
	SState                uint16
	SErrors               uint16
	SMinorRevLevel        uint16
	SLastcheck            uint32
	SCheckinterval        uint32
	SCreatorOs            uint32
	SRevLevel             uint32
	SDefResUID            uint16
	SDefResGID            uint16
	SFirstIno             uint32
	SInodeSize            uint16
	SBlockGroupNr         uint16
	SFeatureCompat        uint32
	SFeatureIncompat      uint32
	SFeatureROCompat      uint32
	SUUID                 [16]byte
	SVolumeName           [16]byte
	SLastMounted          [64]byte
	SAlgorithmUsageBitmap uint32
	SPreallocBlocks       uint8
	SPreallocDirBlocks    uint8
	SReservedGdtBlocks    uint16
	SJournalUUID          [16]byte
	SJournalInum          uint32
	SJournalDev           uint32
	SLastOrphan           uint32
	SHashSeed             [4]uint32
	SDefHashVersion       byte
	SJnlBackupType        byte
	SDefaultMountOpts     uint32
	SFirstMetaBg          uint32
	SMkfsTime             uint32
	SJnlBlocks            [17]uint32
	SBlocksCountHi        uint32
	SRBlocksCountHi       uint32
	SFreeBlocksCountHi    uint32
	SMinExtraIsize        uint16
	SWantExtraIsize       uint16
	SFlags                uint32
	SRaidStride           uint16
	SMmpInterval          uint16
	SMmpBlock             uint64
	SRaidStripeWidth      uint32
	SLogGroupsPerFlex     byte
	SChecksumType         byte
	SReservedPad          uint16
	SKbytesWritten        uint64
	SSnapshotInum         uint32
	SSnapshotId           uint32
	SSnapshotRBlocksCount uint64
	SSnapshotList         uint32
	SErrorCount           uint32
	SFirstErrorTime       uint32
	SFirstErrorIno        uint32
	SFirstErrorBlock      uint64
	SFirstErrorFunc       [32]byte
	SFirstErrorLine       uint32
	SLastErrorTime        uint32
	SLastErrorIno         uint32
	SLastErrorLine        uint32
	SLastErrorBlock       uint64
	SLastErrorFunc        [32]byte
	SMountOpts            [64]byte
	SUsrQuotaInum         uint32
	SGrpQuotaInum         uint32
	SOverheadBlocks       uint32
	SBackupBgs            [2]uint32
	SReserved             [106]uint32
	SChecksum             uint32
}

func LoadSBStruct(file *os.File) *Ext2Sb {
	blockRaw := Load1KB(file, 1024)

	block := (*Ext2Sb)(unsafe.Pointer(&blockRaw[0]))
	return block
}

type Ext2Gd struct {
	BgBlockBitmap     uint32
	BgInodeBitmap     uint32
	BgInodeTable      uint32
	BgFreeBlocksCount uint16
	BgFreeInodesCount uint16
	BgUsedDirsCount   uint16
	BgPad             uint16
	BgReserved        [3]uint32
}

func LoadGDStruct(file *os.File) *Ext2Gd {
	blockRaw := Load1KB(file, 2048)

	block := (*Ext2Gd)(unsafe.Pointer(&blockRaw[0]))
	return block
}

type Ext2Inode struct {
	IMode       uint16
	IUID        uint16
	ISize       uint32
	IAtime      uint32
	ICtime      uint32
	IMtime      uint32
	IDtime      uint32
	IGID        uint16
	ILinksCount uint16
	IBlocks     uint32
	IFlags      uint32
	OSD1        uint32
	IBlock      [15]uint32
	IGeneration uint32
	IFileACL    uint32
	IDirACL     uint32
	IFAddr      uint32
	OSD2        [96]byte
}

func LoadINDTable(file *os.File, gd *BlockGroup, iNodeCount int) []*Ext2Inode {
	ufrom := gd.GroupDescriptors.BgInodeTable * 1024
	from := int(ufrom)
	size := 128
	iNodeTable := make([]*Ext2Inode, iNodeCount)
	for i := 0; i < iNodeCount; i, from = i+1, from+size {
		blockRaw := make([]byte, size)
		file.ReadAt(blockRaw, int64(from))

		block := (*Ext2Inode)(unsafe.Pointer(&blockRaw[0]))
		iNodeTable[i] = block

	}
	return iNodeTable
}

type Ext2DirectoryEntry struct {
	Inode    uint32
	RecLen   uint16
	NameLen  uint8
	FileType uint8
	Name     [255]byte
}

// -- file format --
// EXT2_S_IFSOCK	0xC000	socket
// EXT2_S_IFLNK	0xA000	symbolic link
// EXT2_S_IFREG	0x8000	regular file
// EXT2_S_IFBLK	0x6000	block device
// EXT2_S_IFDIR	0x4000	directory
// EXT2_S_IFCHR	0x2000	character device
// EXT2_S_IFIFO	0x1000	fifo
type Ext2Dentry struct {
	Inode    uint32
	FileType uint8
	Name     string
	RecLen   uint16
}

type Directory struct {
	Dentries []*Ext2Dentry
	Name     string
}

func LoadDentry(file *os.File, from int64, size int) *Ext2Dentry {
	block := LoadNBytes(file, from, size)
	rawDir := (*Ext2DirectoryEntry)(unsafe.Pointer(&block[0]))

	dir := &Ext2Dentry{
		Inode:    rawDir.Inode,
		RecLen:   rawDir.RecLen,
		FileType: rawDir.FileType,
		Name:     string(rawDir.Name[:rawDir.NameLen]),
	}
	return dir
}

func LoadAllDentries(file *os.File, from int64, size int, inodeSize uint32) []*Ext2Dentry {
	entry := LoadDentry(file, from, size)
	dentries := []*Ext2Dentry{entry}
	inSz := int64(inodeSize)
	entriesSize := int64(0)

	for entriesSize < inSz && entry.Inode != 0 {
		recLen := int64(entry.RecLen)
		from += recLen
		entriesSize += recLen
		entry = LoadDentry(file, from, size)
		dentries = append(dentries, entry)
	}

	return dentries
}

func GetInodeAddr(gd *BlockGroup, index int) int64 {
	// Previously it was int64(BOOT_BLOCK_SIZE + (index -1) * BLOCK_SIZE)
	return int64(1024 + (index-1)*1024)
}

func LoadRootDir(file *os.File, gd *BlockGroup) *Directory {
	rInPtr := gd.InodeTable[1]

	dataIndex := GetInodeAddr(gd, int(rInPtr.IBlock[0]))
	dentrySize := int(unsafe.Sizeof(Ext2DirectoryEntry{}))
	rDir := LoadAllDentries(file, dataIndex, dentrySize, rInPtr.ISize)
	rootDir := &Directory{
		Name:     "/",
		Dentries: rDir,
	}

	for i, dir := range rootDir.Dentries {
		fmt.Println(i, "->", fmt.Sprintf("%#v", dir))
	}
	return rootDir
}

func LoadInode(file *os.File, gd *BlockGroup, inodeIndex uint32) *Ext2Inode {
	inodeRaw := LoadNBytes(file, GetInodeAddr(gd, int(inodeIndex)), 128)
	inode := (*Ext2Inode)(unsafe.Pointer(&inodeRaw[0]))
	return inode
}

func LoadInodeBlocks(file *os.File, gd *BlockGroup, inode *Ext2Inode, blockRange int) [][]byte {
	inodeBlocks := [][]byte{}
	for i := 0; i < blockRange; i++ {
		from := GetInodeAddr(gd, int(inode.IBlock[i]))
		block := Load1KB(file, from)
		fmt.Print(string(block))
		inodeBlocks = append(inodeBlocks, block)
	}
	return inodeBlocks
}

// In total there are 15 pointers in the i_block[] array.
// * i_block[0..11] point directly to the first 12 data blocks of the file.
// * i_block[12] points to a single indirect block
// * i_block[13] points to a double indirect block
// * i_block[14] points to a triple indirect block

// Doesn't work
func LoadFile(file *os.File, gd *BlockGroup, fileEntry *Ext2Dentry) {
	size := 1024
	fmt.Println("Data arr", fileEntry.Inode)
	fmt.Println(gd.InodeTable[fileEntry.Inode])
	from := GetInodeAddr(gd, int(fileEntry.Inode))
	blockRaw := make([]byte, size)
	fmt.Println("from", from)
	file.ReadAt(blockRaw, from)
	fmt.Println(string(blockRaw))

	block := (*Ext2Inode)(unsafe.Pointer(&blockRaw[0]))
	// fmt.Println(fmt.Sprintf("%#v", block))

	blockRaw = make([]byte, size)
	fmt.Println(int64(block.IBlock[0]))
	file.ReadAt(blockRaw, int64(block.IBlock[0]))
	fmt.Println(string(blockRaw))

}

func LoadFileO(file *os.File, gd *BlockGroup, fileEntry *Ext2Dentry) {
	inode := LoadInode(file, gd, fileEntry.Inode)
	fmt.Println(fmt.Sprintf("%#v", inode))
	directBlocks := LoadInodeBlocks(file, gd, inode, 12)

	singleIndirect := LoadInode(file, gd, inode.IBlock[12])
	snInBlocks := LoadInodeBlocks(file, gd, singleIndirect, 15)

	doubleIndirect := LoadInode(file, gd, inode.IBlock[13])
	dbInBlocks := [][]byte{}
	for i := 0; i < 15; i++ {
		dbInNode := LoadInode(file, gd, doubleIndirect.IBlock[i])
		dbInBlocks = append(dbInBlocks, LoadInodeBlocks(file, gd, dbInNode, 15)...)
	}

	tripleIndirect := LoadInode(file, gd, inode.IBlock[14])
	trInBlocks := [][]byte{}
	for i := 0; i < 15; i++ {
		for j := 0; j < 15; j++ {
			trInNode := LoadInode(file, gd, tripleIndirect.IBlock[j])
			trInBlocks = append(trInBlocks, LoadInodeBlocks(file, gd, trInNode, 15)...)
		}
	}

	var fileBuffer bytes.Buffer
	fileContents := [][]byte{}
	fileContents = append(fileContents, directBlocks...)
	fileContents = append(fileContents, snInBlocks...)
	fileContents = append(fileContents, dbInBlocks...)
	fileContents = append(fileContents, trInBlocks...)

	for _, bf := range fileContents {
		fileBuffer.Write(bf)
	}
	fmt.Println(fileBuffer.String())
}
