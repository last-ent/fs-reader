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
	if err != nil {
		fmt.Println(err)
		return
	}
	// b1 := make([]byte, 1)
	// b2 := make([]byte, 1)
	// file.ReadAt(b1, 1080)
	// file.ReadAt(b2, 1081)
	// fmt.Println(fmt.Sprintf("%x %x", b1, b2))

	// Read super block byte stream
	groupDesc := ext.LoadBlockGroup(file)
	fmt.Println(fmt.Sprintf("%+v", groupDesc))

}
