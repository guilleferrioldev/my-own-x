package main

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
)

func gitWriteTree(args []string) {
	if len(args) != 2 {
		printUsageAndExit("write-tree \n")
		return
	}

	cwd, _ := os.Getwd()
	fmt.Printf("%x\n", writeTree(cwd))
}

type treeEntry struct {
	name string
	mode string
	hash []byte
}

func writeTree(path string) []byte {
	dir, err := os.Open(path)
	if err != nil {
		fatal(err.Error())
	}
	defer dir.Close()

	if info, _ := dir.Stat(); !info.IsDir() {
		fatal("not a directory: %s\n", path)
	}

	entries, err := dir.ReadDir(0)
	if err != nil {
		fatal(err.Error())
	}

	treeEntries := []*treeEntry{}

	for _, entry := range entries {
		if entry.Name() == myGitDir {
			continue
		}
		te := new(treeEntry)
		te.name = entry.Name()
		fullPath := filepath.Join(path, te.name)
		if entry.IsDir() {
			te.mode = "40000" // directory
			te.hash = writeTree(fullPath)
		} else {
			// 100755 (executable file)
			// 120000 (symbolic link)
			te.mode = "100644" // regular file
			te.hash = hashFile(true, fullPath)
		}
		treeEntries = append(treeEntries, te)
	}

	slices.SortFunc(treeEntries, func(a, b *treeEntry) int {
		return cmp.Compare(a.name, b.name)
	})

	content := []byte{}
	for _, entry := range treeEntries {
		content = append(content, []byte(entry.mode)...)
		content = append(content, ' ')
		content = append(content, []byte(entry.name)...)
		content = append(content, '\000')
		content = append(content, entry.hash...)
	}

	return hashObject(true, "tree", int64(len(content)), content)
}
