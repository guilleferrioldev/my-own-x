package main

import (
	"bufio"
	"bytes"
	"cmp"
	"compress/zlib"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
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
		case "clone":
			gitClone(args)
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

func gitInit() {
	cwd, err := os.Getwd()
	if err != nil {
		fatal("Error getting current working directory: %v", err)
	}

	repoPath := filepath.Join(cwd, myGitDir)

	cmd := exec.Command("mkdir", "-p", filepath.Join(repoPath, "objects"), filepath.Join(repoPath, "refs"))
	cmd.Dir = cwd
	err = cmd.Run()
	if err != nil {
		fatal("Error creating .mygit directories: %v", err)
	}

	headPath := filepath.Join(repoPath, "HEAD")
	headContent := []byte("ref: refs/heads/master\n")

	err = os.WriteFile(headPath, headContent, 0644)
	if err != nil {
		fatal("Error writing to HEAD file: %v", err)
	}

	fmt.Printf("Initialized empty Git repository in %s\n", repoPath)
}

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

func gitListTree(args []string) {
	if len(args) < 3 || (len(args) == 4 && args[2] != "--name-only" && args[2] != "--object-only" && args[2] != "-l") {
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

func gitCommitTree(args []string) {
	usage := false
	if len(args) < 3 {
		usage = true
	}

	var treeHash, parentHash, message string
	for i := 0; !usage && i < len(args); i++ {
		switch args[i] {
		case "-p":
			if parentHash == "" && i+1 < len(args) {
				parentHash = args[i+1]
			} else {
				usage = true
			}
			i++ // skip
		case "-m":
			if message == "" && i+1 < len(args) {
				message = args[i+1]
			} else {
				usage = true
			}
			i++ // skip
		default:
			treeHash = args[i]
		}
	}

	if usage {
		printUsageAndExit("commit-tree <tree_sha> [-p <parent_sha>] [-m <message>] \n")
		return
	}

	// make sure tree_sha and parent_sha exists and have the right type
	treeType, _ := getObjTypeAndSize(treeHash)
	if treeType != "tree" {
		fatal("expected '%s' to be a 'tree' object, got: %s", treeHash, treeType)
	}

	parentType, _ := getObjTypeAndSize(treeHash)
	if parentType != "tree" {
		fatal("expected '%s' to be a 'tree' object, got: %s", parentHash, parentType)
	}

	// TODO: Using "git config" itself to get the config for now...
	username, email := getGitConfig("user.name"), getGitConfig("user.email")
	now := time.Now()
	timestamp := now.Unix()
	_, tzOffset := now.Zone()
	tzHours, tzMinutes := tzOffset/3600, (tzOffset/60)%60
	timezone := tzHours*100 + tzMinutes

	content := fmt.Sprintf("tree %s\n", treeHash)
	if parentHash != "" {
		content += fmt.Sprintf("parent %s\n", parentHash)
	}
	content += fmt.Sprintf("author %s <%s> %d %+05d\n", username, email, timestamp, timezone)
	content += fmt.Sprintf("committer %s <%s> %d %+05d\n", username, email, timestamp, timezone)
	content += fmt.Sprintf("\n%s\n", message)

	commitHash := hashObject(true, "commit", int64(len(content)), []byte(content))
	fmt.Printf("%x\n", commitHash)
}

func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", key)
	output, err := cmd.Output()
	if err != nil {
		fatal(err.Error())
	}
	return strings.TrimRight(string(output), "\r\n")
}

func gitClone(args []string) {
	fmt.Println(args)
	if len(args) < 4 {
		printUsageAndExit("clone <repo> <dir> \n")
		return
	}

	repoUrl := args[2]
	directory := args[3]

	// TODO: refactor gitInit to accept parameter to new directory
	// TEMP: make new directory and initialize .mygit
	os.Mkdir(directory, 0755)
	os.Chdir(directory)
	gitInit()

	packContent, head := fetchGitPack(repoUrl)

	os.MkdirAll(filepath.Join(myGitDir, "refs", "heads"), 0755)
	os.WriteFile(filepath.Join(myGitDir, "refs", "heads", "master"), []byte(head), 0644)

	unpackObjects(packContent)
	// "checkout" files to workdir
	headHash, _ := hex.DecodeString(string(head))
	checkoutCommit(headHash)
}

func fetchGitPack(repoUrl string) (packContent []byte, head string) {
	respGet, err := http.Get(repoUrl + "/info/refs?service=git-upload-pack")
	if err != nil {
		fatal(err.Error())
	}
	defer respGet.Body.Close()

	if respGet.StatusCode != 200 {
		fatal("could not fetch %q - status code: %d", repoUrl, respGet.StatusCode)
	}

	contentType := respGet.Header.Get("Content-Type")
	if contentType != "application/x-git-upload-pack-advertisement" {
		fatal("unexpected content type: %q", contentType)
	}

	refs := map[string]string{}

	// start parsing the pack response
	sizeBuffer := make([]byte, 4)
	for {
		_, err := respGet.Body.Read(sizeBuffer)
		if err != nil {
			if err == io.EOF {
				break
			}
			fatal(err.Error())
		}

		size, err := strconv.ParseUint(string(sizeBuffer), 16, 32)
		if err != nil {
			fatal(err.Error())
		}

		if size > 4 {
			dataBuffer := make([]byte, size-4)
			_, err = respGet.Body.Read(dataBuffer)
			if err != nil {
				fatal(err.Error())
			}

			fmt.Printf("size=%d data=%q\n", size, string(dataBuffer))

			// strip newline
			data := string(dataBuffer)
			if data[len(data)-1] == '\n' {
				data = data[:len(data)-1]
			}

			if data[0] == '#' {
				if data != "# service=git-upload-pack" {
					fatal("unexpected header: %q", data)
				}
			} else {
				hash := data[:40]
				// ignore whitespace
				parts := strings.Split(data[41:], "\000")
				ref := parts[0]
				if len(parts) > 1 {
					capabilities := strings.Split(parts[1], " ")
					for _, capability := range capabilities {
						fmt.Printf("capability: %s\n", capability)
					}
				}
				fmt.Printf("hash=%s ref=%s\n", hash, ref)
				refs[ref] = hash
			}
		} else {
			fmt.Printf("size=%d (terminator)\n", size)
		}
	}

	for ref, hash := range refs {
		fmt.Println(ref, hash)
	}

	var ok bool
	if head, ok = refs["HEAD"]; !ok {
		fatal("no HEAD reference found")
	}

	fmt.Println("====")

	postHeader := "application/x-git-upload-pack-request"
	postBody := fmt.Sprintf("0032want %s\n00000009done\n", refs["HEAD"])
	respPost, err := http.Post(repoUrl+"/git-upload-pack", postHeader, strings.NewReader(postBody))
	if err != nil {
		fatal(err.Error())
	}
	defer respPost.Body.Close()

	if respGet.StatusCode != 200 {
		fatal("could not fetch %q - status code: %d", repoUrl, respGet.StatusCode)
	}

	nakExpected := []byte("0008NAK\n")
	nakHeader := make([]byte, 8)
	_, err = respPost.Body.Read(nakHeader)
	if err != nil {
		fatal(err.Error())
	}
	if slices.Compare(nakExpected, nakHeader) != 0 {
		fatal("unexpected header on response. got: %q - want: %q\n", nakHeader, nakExpected)
	}

	packContent, err = io.ReadAll(respPost.Body)
	if err != nil {
		fatal(err.Error())
	}
	return
}

const OBJ_COMMIT = 1
const OBJ_TREE = 2
const OBJ_BLOB = 3
const OBJ_TAG = 4
const OBJ_OFS_DELTA = 6
const OBJ_REF_DELTA = 7

type objDelta struct {
	source  []byte
	content []byte
}

func unpackObjects(packContent []byte) {
	reader := bufio.NewReader(bytes.NewReader(packContent))

	// begin reading the pack file now
	packHeader := make([]byte, 12)
	_, err := reader.Read(packHeader)
	if err != nil {
		fatal(err.Error())
	}
	if slices.Compare(packHeader[:4], []byte("PACK")) != 0 {
		fatal("invalid PACK header")
	}
	version := bigEndianBytesToUint(packHeader[4:8])
	objCount := bigEndianBytesToUint(packHeader[8:12])
	fmt.Println(version, objCount)

	// save deltas to apply after unpacking
	savedObjDeltas := []objDelta{}

	// extracting objects from pack file received

	offset := uint64(12)
	for index := 0; index < int(objCount); index++ {

		var objRefDeltaHash []byte

		value, err := reader.ReadByte()
		lenghtBytesRead := uint64(1)
		if err != nil {
			fatal(err.Error())
		}

		objType := (value >> 4) & 0b00000111
		informedSize := uint64(value & 0b00001111)
		shift := 4
		for value&0b10000000 != 0 {
			value, err = reader.ReadByte()
			lenghtBytesRead++
			if err != nil {
				fatal(err.Error())
			}
			informedSize = informedSize | uint64(value&0b01111111)<<shift
			if value&0b10000000 == 0 {
				break
			}
			shift += 7
		}

		fmt.Printf("index=%2d\ttype=%d\toffset=%5d\tsize=%5d", index, objType, offset, informedSize)
		if objType == OBJ_OFS_DELTA {
			fatal("OBJ_OFS_DELTA not implemented yet!")
		} else if objType == OBJ_REF_DELTA {
			objRefDeltaHash = make([]byte, 20)
			reader.Read(objRefDeltaHash)
		}
		// NOTE: no idea why reported size is too small in some cases...
		if informedSize < 1024 {
			informedSize = 1024
		}
		compressedBuffer, err := reader.Peek(int(informedSize))
		if err != nil && err != io.EOF {
			fatal(err.Error())
		}
		bytesBuffer := bytes.NewReader(compressedBuffer)
		zreader, err := zlib.NewReader(bytesBuffer)
		if err != nil {
			fatal(err.Error())
		}
		uncompressedBuffer := make([]byte, informedSize)
		actualSize, err := zreader.Read(uncompressedBuffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			fatal(err.Error())
		}

		compressedBytesRead := uint64(bytesBuffer.Size() - int64(bytesBuffer.Len()))
		fmt.Printf("\tactual_size=%5d\tcompressed_size=%5d", uint64(actualSize), +compressedBytesRead)

		if objType == OBJ_REF_DELTA {
			fmt.Printf("\tobjRefDelta=%x", objRefDeltaHash)
		}

		fmt.Printf("\tcontent=%q\n", uncompressedBuffer[:actualSize])

		// hash objects and write objects that were read from the pack file

		switch objType {
		case OBJ_BLOB:
			hashObject(true, "blob", int64(actualSize), uncompressedBuffer[:actualSize])
		case OBJ_TREE:
			hashObject(true, "tree", int64(actualSize), uncompressedBuffer[:actualSize])
		case OBJ_COMMIT:
			hashObject(true, "commit", int64(actualSize), uncompressedBuffer[:actualSize])
		case OBJ_TAG:
			hashObject(true, "tag", int64(actualSize), uncompressedBuffer[:actualSize])
		case OBJ_REF_DELTA:
			savedObjDeltas = append(savedObjDeltas, objDelta{objRefDeltaHash, uncompressedBuffer[:actualSize]})
		}

		reader.Discard(int(compressedBytesRead))
		offset += uint64(lenghtBytesRead + compressedBytesRead)
	}

	// applying deltas
	// reference: https://codewords.recurse.com/issues/three/unpacking-git-packfiles#applying-deltas

	for _, savedObjDelta := range savedObjDeltas {
		sourceHash, content := savedObjDelta.source, savedObjDelta.content
		fmt.Printf("parent=%x\tcontent=%v\n", sourceHash, content)

		i := 0
		sourceSize := uint64(content[i]) & 0b01111111
		shift := 7
		for content[i]&0b10000000 != 0 {
			i++
			sourceSize |= uint64(content[i]&0b01111111) << shift
			shift += 7
		}
		fmt.Printf("sourceSize=%d\t", sourceSize)
		i++

		targetSize := uint64(content[i]) & 0b01111111
		shift = 7
		for content[i]&0b10000000 != 0 {
			i++
			targetSize |= uint64(content[i]&0b01111111) << shift
			shift += 7
		}
		fmt.Printf("targetSize=%d\n", targetSize)
		i++

		targetBuffer := make([]byte, targetSize)

		sourceType, readSize, sourceBuffer := readObject(sourceHash)
		// TODO: are other types allowed here?
		if sourceSize != readSize {
			fatal("unexpected source size for delta: got %d - want %d\n", readSize, sourceSize)
		}

		var sourceIndex, targetIndex uint32

		// iterate on instructions (copy/insert)
		for i < len(content) {
			op := content[i] >> 7
			if op == 1 { // copy operation

				// decode which bytes to read that are non-zero from the copy instruction
				lengthBitmap := content[i] >> 4 & 0b0000111
				offsetBitmap := content[i] & 0b00001111
				i++

				offset, length := uint32(0), uint32(0)

				if offsetBitmap&0b0001 != 0 {
					offset |= uint32(content[i])
					i++
				}
				if offsetBitmap&0b0010 != 0 {
					offset |= uint32(content[i]) << 8
					i++
				}
				if offsetBitmap&0b0100 != 0 {
					offset |= uint32(content[i]) << 16
					i++
				}
				if offsetBitmap&0b1000 != 0 {
					offset |= uint32(content[i]) << 24
					i++
				}

				if lengthBitmap&0b0001 != 0 {
					length |= uint32(content[i])
					i++
				}
				if lengthBitmap&0b0010 != 0 {
					length |= uint32(content[i]) << 8
					i++
				}
				if lengthBitmap&0b0100 != 0 {
					length |= uint32(content[i]) << 16
					i++
				}

				sourceIndex = offset
				fmt.Printf("copying %d:%d into %d:%d\n", sourceIndex, sourceIndex+length, targetIndex, targetIndex+length)
				copy(targetBuffer[targetIndex:targetIndex+length], sourceBuffer[sourceIndex:sourceIndex+length])
				targetIndex += length
			} else { // insert operation
				length := uint32(content[i]) // length is on the operation itself 0-127
				i++
				fmt.Printf("inserting %d bytes into %d:%d\n", length, targetIndex, targetIndex+length)
				copy(targetBuffer[targetIndex:targetIndex+length], content[i:i+int(length)])
				targetIndex += length
				i += int(length)
			}
		}

		targetHash := hashObject(true, sourceType, int64(targetSize), targetBuffer)
		fmt.Printf("delta applied source: %x target: %x\n", sourceHash, targetHash)
	}
}

func bigEndianBytesToUint(b []byte) uint {
	return uint(b[0])<<24 | uint(b[1])<<16 | uint(b[2])<<8 | uint(b[3])
}

func readObject(hash []byte) (objType string, objSize uint64, content []byte) {

	objPath := filepath.Join(myGitDir, "objects", fmt.Sprintf("%x", hash[:1]), fmt.Sprintf("%x", hash[1:]))
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

	lengthStr, err := reader.ReadString(0)
	lengthStr = lengthStr[:len(lengthStr)-1]
	if err != nil {
		fatal(err.Error())
	}
	objSize, _ = strconv.ParseUint(lengthStr, 10, 64)

	content = make([]byte, objSize)
	_, err = reader.Read(content)
	if err != nil && err != io.EOF {
		fatal(err.Error())
	}
	return
}

func checkoutCommit(head []byte) {
	objType, _, content := readObject(head)
	if objType != "commit" {
		fatal("expected 'commit' type, got: %s\n", objType)
	}
	reader := bufio.NewReader(bytes.NewReader(content))
	reader.ReadString(' ')
	treeHash, _ := reader.ReadString('\n')
	hash, _ := hex.DecodeString(treeHash)

	// TODO: make sure directory is empty
	root, err := os.Getwd()
	if err != nil {
		fatal(err.Error())
	}
	checkoutTree(hash, root)
}

func checkoutTree(tree []byte, path string) {
	fmt.Printf("dir: %s\n", path)

	os.MkdirAll(path, 0755)

	objType, _, content := readObject(tree)
	if objType != "tree" {
		fatal("expected 'tree' type, got: %s\n", objType)
	}
	reader := bufio.NewReader(bytes.NewReader(content))

	for {
		fileMode, err := reader.ReadString(' ')
		if err != nil {
			if err == io.EOF {
				break
			}
			fatal(err.Error())
		}
		fileMode = "000000" + fileMode[:len(fileMode)-1]
		fileMode = fileMode[len(fileMode)-6:]

		name, err := reader.ReadString('\000')
		if err != nil {
			fatal(err.Error())
		}
		name = name[:len(name)-1]

		hash := make([]byte, 20)
		_, err = reader.Read(hash)
		if err != nil {
			fatal(err.Error())
		}

		switch fileMode {
		case "100644":
			checkoutFile(hash, filepath.Join(path, name))
		case "040000":
			checkoutTree(hash, filepath.Join(path, name))
		default:
			fmt.Printf("unknown file mode: %s skipping: %s (%x)\n", fileMode, name, hash)
		}
	}
}

func checkoutFile(hash []byte, path string) {
	fmt.Printf("file: %s\n", path)

	objType, _, content := readObject(hash)
	if objType != "blob" {
		fatal("expected 'blob' type, got: %s\n", objType)
	}

	file, err := os.Create(path)
	if err != nil {
		fatal(err.Error())
	}
	defer file.Close()

	_, err = file.Write(content)
	if err != nil {
		fatal(err.Error())
	}
}
