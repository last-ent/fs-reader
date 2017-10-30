package ext

import (
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

func LoadBlockGroup(file *os.File) *BlockGroup {
	blockGroup := BlockGroup{
		SuperBlock:       LoadSBStruct(file), // 1024 bootblock, 1024*2
		GroupDescriptors: LoadGDStruct(file), // 2048...1024*3
		DataBlockBitmap:  Load1KB(file, 1024*4),
		InodeBitmap:      Load1KB(file, 1024*5),
	}
	INodesCount := int(blockGroup.SuperBlock.SBlocksPerGroup / blockGroup.SuperBlock.SInodesPerGroup)
	fmt.Println("Inodes Count", blockGroup.SuperBlock.SInodesCount, "Blocks Count", blockGroup.SuperBlock.SBlocksCount)
	fmt.Println("Inodes per group", blockGroup.SuperBlock.SInodesPerGroup, "Blocks per group", blockGroup.SuperBlock.SBlocksPerGroup)

	// INodesCount := int(blockGroup.SuperBlock.SInodesPerGroup) / int(iNodesPerBlock)
	fmt.Println(INodesCount)

	blockGroup.InodeTable = LoadINDTable(
		file,
		blockGroup.GroupDescriptors.BgInodeTable,
		INodesCount,
	)
	return &blockGroup
}

func Load1KB(file *os.File, from int64) []byte {
	size := 1024
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
	size := 1024
	var from int64 = 1024
	blockRaw := make([]byte, size)
	file.ReadAt(blockRaw, from)

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
	size := 1024
	var from int64 = 2048
	blockRaw := make([]byte, size)
	file.ReadAt(blockRaw, from)

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

func LoadINDTable(file *os.File, ufrom uint32, iNodeCount int) []*Ext2Inode {
	from := int(ufrom)
	size := 128
	iNodeTable := make([]*Ext2Inode, iNodeCount)
	for i := 0; i < iNodeCount; i, from = i+1, from+size {
		blockRaw := make([]byte, size)
		file.ReadAt(blockRaw, int64(from))

		block := (*Ext2Inode)(unsafe.Pointer(&blockRaw[0]))
		// iNodeTable = append(iNodeTable, block)
		iNodeTable[i] = block
	}
	fmt.Println(iNodeCount, len(iNodeTable))
	return iNodeTable
}
