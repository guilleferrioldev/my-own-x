package main

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

func gitListTree(args []string) {
	if len(args) < 4 || (args[2] != "--name-only" && args[2] != "--object-only" && args[2] != "-l") {
		printUsageAndExit("ls-tree [(-l | --name-only | --object-only)] <tree_sha> \n")
		return
	}

	var nameOnly, objectOnly, longFormat bool
	var objName string
	if args[2] == "--name-only" {
		objName = args[3]
		nameOnly = true
	} else if args[2] == "--object-only" {
		objName = args[3]
		objectOnly = true
	} else if args[2] == "-l" {
		objName = args[3]
		longFormat = true
	} else {
		objName = args[2]
	}

	if len(objName) != 40 {
		fmt.Printf("error: Not a valid object name %s\n", objName)
		return
	}

	objDir := filepath.Join(myGitDir, "objects", objName[:2])
	info, err := os.Stat(objDir)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	if !info.IsDir() {
		fmt.Printf("error: not a directory %s\n", objDir)
		return
	}

	objPath := filepath.Join(objDir, objName[2:])
	file, err := os.Open(objPath)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	defer file.Close()

	zipReader, err := zlib.NewReader(file)
	if err != nil {
		fmt.Printf(err.Error())
		return
	}

	reader := bufio.NewReader(zipReader)
	objType, _ := reader.ReadString(' ')
	objType = objType[:len(objType)-1]

	if objType != "tree" {
		fmt.Printf("expected a 'tree' node, found: %q\n", objType)
		return
	}

	lengthStr, err := reader.ReadString(0)
	lengthStr = lengthStr[:len(lengthStr)-1]
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	objSize, _ := strconv.ParseInt(lengthStr, 10, 64)

	if objSize == 0 {
		fmt.Printf("error: object file %s is empty", objPath)
		return
	}

	hash := make([]byte, 20)
	for {
		fileMode, err := reader.ReadString(' ')
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf(err.Error())
			return
		}
		fileMode = "000000" + fileMode[:len(fileMode)-1]
		fileMode = fileMode[len(fileMode)-6:]

		name, err := reader.ReadString('\000')
		if err != nil {
			fmt.Printf(err.Error())
			return
		}
		name = name[:len(name)-1]

		_, err = reader.Read(hash)
		if err != nil {
			fmt.Printf(err.Error())
			return
		}

		if nameOnly {
			fmt.Println(name)
		} else if objectOnly {
			fmt.Printf("%x\n", hash)
		} else if longFormat {
			objType, objSize := getObjTypeAndSize(fmt.Sprintf("%x", hash))
			fmt.Printf("%s %s %x\t%d\t%s\n", fileMode, objType, hash, objSize, name)
		} else {
			objType, _ := getObjTypeAndSize(fmt.Sprintf("%x", hash))
			fmt.Printf("%s %s %x\t%s\n", fileMode, objType, hash, name)
		}
	}
}
