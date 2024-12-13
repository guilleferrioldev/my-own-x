package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func handleFileRequest(connection net.Conn, path string) (statusCode int, statusMessage string) {
	info, err := os.Stat("./files/" + path)
	if err != nil {
		if os.IsNotExist(err) {
			return 404, "Not Found"
		}
		return 500, "Internal Server Error"
	}

	if info.IsDir() {
		return 500, "Internal Server Error"
	}

	file, err := os.Open("./files/" + path)
	if err != nil {
		return 500, "Internal Server Error"
	}
	defer file.Close()

	size, _ := file.Seek(0, io.SeekEnd)
	statusCode, statusMessage = 200, "OK"
	httpHeader := fmt.Sprintf("HTTP/1.1 %d %s\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n",
		statusCode, statusMessage, size)
	_, err = connection.Write([]byte(httpHeader))
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
	}

	file.Seek(0, io.SeekStart)
	data := make([]byte, size)
	_, err = file.Read(data)
	if err != nil {
		return 500, "Internal Server Error"
	}

	_, err = connection.Write(data)
	if err != nil {
		fmt.Printf("Error sending response: %v\n", err)
	}
	fmt.Printf("Served file %s to client\n", "./files/"+path)

	return
}
