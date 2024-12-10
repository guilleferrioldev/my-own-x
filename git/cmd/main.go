package main

import (
	"bufio"
	"compress/zlib"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	myGitDir = ".mygit"
)

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("Welcome to MyGit!")

	for {
		fmt.Print("> ")
		input, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			break
		}

		input = strings.TrimSpace(input)
		if input == "exit" {
			break
		}

		args := strings.Fields(input)
		if len(args) < 1 {
			fmt.Println("You must specify a command.")
			continue
		}

		if args[0] != "git" {
			fmt.Println("You must specify a valid command.")
			continue
		}

		command := args[1]
		switch command {
		case "init":
			gitInit()
		case "cat-file":
			gitCatFile(args)
		case "hash-object":
			gitHashObject(args)
		case "ls-tree":
			gitListTree(args)
		case "write-tree":
			gitWriteTree(args)
		case "commit-tree":
			gitCommitTree(args)
		default:
			fmt.Printf("Comando inválido: %s\n", command)
		}
	}

	fmt.Println("Leaving...")
}

func printUsageAndExit(command string) {
	myName := filepath.Base(os.Args[0])
	if len(command) == 0 {
		fmt.Printf("usage: %s <command> [<args>...]", myName)
	} else {
		fmt.Printf("usage: %s %s", myName, command)
	}
}

func fatal(msg string, args ...any) {
	fmt.Fprintf(os.Stderr, msg, args...)
	os.Exit(128)
}

func getObjTypeAndSize(objName string) (objType string, objSize int64) {
	objPath := filepath.Join(myGitDir, "objects", objName[:2], objName[2:])

	file, err := os.Open(objPath)
	if err != nil {
		fatal(err.Error())
	}
	defer file.Close()

	zipReader, err := zlib.NewReader(file)
	if err != nil {
		fatal(err.Error())
	}

	reader := bufio.NewReader(zipReader)
	objType, _ = reader.ReadString(' ')
	objType = objType[:len(objType)-1]

	lengthStr, _ := reader.ReadString(0)
	lengthStr = lengthStr[:len(lengthStr)-1]
	objSize, _ = strconv.ParseInt(lengthStr, 10, 64)

	return
}
