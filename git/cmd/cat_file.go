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

func gitCatFile(args []string) {
	if len(args) < 4 || !(args[2] == "-p" || args[2] == "-t" || args[2] == "-s" || args[2] == "-e") {
		printUsageAndExit("cat-file (-p | -t | -s | -e) <object> \n")
		return
	}

	objName := args[3]
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

	lengthStr, err := reader.ReadString(0)
	lengthStr = lengthStr[:len(lengthStr)-1]
	if err != nil {
		fmt.Printf(err.Error())
		return
	}
	objSize, _ := strconv.ParseInt(lengthStr, 10, 64)

	switch args[2] {
	case "-p":
		// TODO: default action "-p" (pretty-print)
	case "-t":
		fmt.Println(objType)
	case "-s":
		fmt.Println(objSize)

		if objSize == 0 {
			fmt.Printf("error: object file %s is empty", objPath)
		}
	default:
		io.Copy(os.Stdout, reader)
		return
	}
}
