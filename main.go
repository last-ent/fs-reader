package main

import (
	"fmt"
	"os"

	"github.com/last-ent/fs-reader/ext"
)

const ext2Magic uint16 = 0xef53

// os.File.Seek :: whence
// 0 -> Relative to start of file
// 1 -> Relative to current offset
// 2 -> Relative to end of file

func main() {
	file, err := os.Open("/home/entux/Documents/Code/fsfs/linux.ex2")
	defer file.Close()
	if err != nil {
		fmt.Println(err)
		return
	}
	// Read super block byte stream
	groupDesc := ext.LoadBlockGroup(file)
	ext.LoadRootDir(file, groupDesc)
}
