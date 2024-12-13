package main

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

func handleConnection(connection net.Conn, directory string) {
	defer connection.Close()

	var requestMethod, requestPath, requestVersion string
	bytesRead, _ := fmt.Fscanf(connection, "%s %s %s\r\n", &requestMethod, &requestPath, &requestVersion)
	fmt.Printf("Read %d bytes from client\n", bytesRead)

	var statusCode int
	var statusMessage, responseBody, requestFilePath, extraHeaders string
	if requestVersion != "HTTP/1.1" {
		statusCode, statusMessage = 400, "Bad Request"
	} else {
		statusCode, statusMessage = 200, "OK"
		if requestMethod != "GET" && requestMethod != "POST" {
			statusCode, statusMessage = 501, "Not Implemented"
		} else if requestPath == "/" {
			// do nothing
		} else if requestPath == "/user-agent" {
			scanner := bufio.NewScanner(connection)
			for scanner.Scan() {
				line := scanner.Text()
				if strings.HasPrefix(line, "User-Agent: ") {
					responseBody = line[12:]
					break
				}
			}
		} else if strings.HasPrefix(requestPath, "/echo/") {
			responseBody = requestPath[6:]
			scanner := bufio.NewScanner(connection)
			var acceptEncodings string
			fmt.Println("processing headers...")
			for scanner.Scan() {
				line := scanner.Text()
				fmt.Println(line)
				if strings.HasPrefix(line, "Accept-Encoding: ") {
					acceptEncodings = line[17:]
				} else if line == "" {
					break
				}
			}
			fmt.Println("headers processed!")
			for _, acceptEncoding := range strings.Split(acceptEncodings, ",") {
				if strings.Trim(acceptEncoding, " ") == "gzip" {
					extraHeaders = "Content-Encoding: gzip\r\n"
					buf := bytes.NewBuffer([]byte{})
					writer := gzip.NewWriter(buf)
					writer.Write([]byte(responseBody))
					writer.Close()
					responseBody = buf.String()
					break
				}
			}
		} else if strings.HasPrefix(requestPath, "/files/") {
			requestFilePath = requestPath[7:]
			fullFilePath := filepath.Join(directory, requestFilePath)
			switch requestMethod {
			case "GET":
				statusCode, statusMessage = handleFileRequest(connection, fullFilePath)
				if statusCode == 200 {
					return
				}
			case "POST":
				statusCode, statusMessage = handleFileUpload(connection, fullFilePath)
			}
		} else {
			statusCode, statusMessage = 404, "Not Found"
		}
	}

	fmt.Println(statusCode, responseBody)
	httpResponse := fmt.Sprintf("HTTP/1.1 %d %s\r\n%sContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s\r\n",
		statusCode, statusMessage, extraHeaders, len(responseBody), responseBody)
	bytesSent, err := connection.Write([]byte(httpResponse))
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Sent %d bytes to client (expected: %d)\n", bytesSent, len(httpResponse))

}
