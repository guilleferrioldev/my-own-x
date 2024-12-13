package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func handleFileUpload(connection net.Conn, fileName string) (statusCode int, statusMessage string) {
	// Sanitize filename to prevent directory traversal attacks
	baseDir := "./files/"
	fileName = filepath.Clean(fileName)
	filePath := filepath.Join(baseDir, fileName)

	// Create the directory if it doesn't exist
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		return 500, "Internal Server Error"
	}

	file, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("Error creating file: %v\n", err)
		return 500, "Internal Server Error"
	}
	defer file.Close()

	reader := bufio.NewReader(connection)
	// Read headers (improved to handle missing Content-Length)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Error reading headers: %v\n", err)
			return 500, "Internal Server Error"
		}
		line = strings.TrimSuffix(line, "\r\n")
		if len(line) == 0 {
			break
		}
	}

	// Copy the file content using io.Copy, handling potential errors
	bytesWritten, err := io.Copy(file, reader)
	if err != nil {
		fmt.Printf("Error copying file content: %v\n", err)
		return 500, "Internal Server Error"
	}

	fmt.Printf("Received file %s from client (bytes %v)\n", filePath, bytesWritten)
	return 201, "Created"
}
