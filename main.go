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
	fmt.Println(fmt.Sprintf("%#v", groupDesc))
	rootDir := ext.LoadRootDir(file, groupDesc)
	var asdf *ext.Ext2Dentry
	i := 0
	for ; i < len(rootDir.Dentries); i++ {
		asdf = rootDir.Dentries[i]
		if asdf.Name == "asdf.txt" {
			break
		}
	}

	if i == len(rootDir.Dentries) {
		fmt.Println("Not found")
		return
	}

	ext.LoadFile(file, groupDesc, asdf)

}
