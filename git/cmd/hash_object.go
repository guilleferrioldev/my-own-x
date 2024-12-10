package main

import (
	"compress/zlib"
	"crypto/sha1"
	"fmt"
	"os"
	"path/filepath"
)

func gitHashObject(args []string) {
	if len(args) < 3 || (args[2] == "-w" && len(args) < 4) {
		printUsageAndExit("hash-object [-w] <object> \n")
		return
	}

	var writeObject bool
	var filename string
	if args[2] == "-w" {
		filename = args[3]
		writeObject = true
	} else {
		filename = args[2]
	}

	fmt.Printf("%x\n", hashFile(writeObject, filename))
}

func hashFile(writeObject bool, filename string) []byte {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf(err.Error())
		return nil
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		fmt.Printf(err.Error())
		return nil
	}
	if info.IsDir() {
		fmt.Printf("'%s' is a directory", info.Name())
		return nil
	}
	fileSize := info.Size()

	content := make([]byte, fileSize)
	_, err = file.Read(content)
	if err != nil {
		fmt.Printf(err.Error())
		return nil
	}

	return hashObject(writeObject, "blob", fileSize, content)
}

func hashObject(writeObject bool, contentType string, contentSize int64, content []byte) []byte {
	payload := []byte(fmt.Sprintf("%s %d\000", contentType, contentSize))

	s := sha1.New()
	s.Write(payload)
	s.Write(content)

	hash := s.Sum(nil)
	objName := fmt.Sprintf("%x", hash)

	if !writeObject {
		return hash
	}

	objDir := filepath.Join(myGitDir, "objects", objName[:2])
	objPath := filepath.Join(objDir, objName[2:])

	if fileExists(objPath) {
		return hash
	}

	err := os.MkdirAll(objDir, 0755)
	if err != nil {
		fmt.Printf(err.Error())
		return nil
	}

	objFile, err := os.OpenFile(objPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return nil
	}
	defer objFile.Close()

	writer := zlib.NewWriter(objFile)
	_, err = writer.Write(payload)
	if err != nil {
		fmt.Printf("Error writing payload: %v\n", err)
		return nil
	}
	_, err = writer.Write(content)
	if err != nil {
		fmt.Printf("Error writing content: %v\n", err)
		return nil
	}

	err = writer.Close()
	if err != nil {
		fmt.Printf("Error closing zlib writer: %v\n", err)
		return nil
	}
	return hash
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fatal(err.Error())
	}
	return true
}
