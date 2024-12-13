package main

import (
	"flag"
	"fmt"
	"net"
	"os"
)

func main() {
	var directory, host string
	var port int

	flag.StringVar(&host, "host", "0.0.0.0", "interface ip/host")
	flag.IntVar(&port, "port", 4221, "tcp port to listen for connections")
	flag.StringVar(&directory, "directory", ".", "directory from which to serve files")
	flag.Parse()

	info, err := os.Stat(directory)
	if err != nil {
		fmt.Printf("Failed to check directory path: %v\n", err)
		os.Exit(1)
	}
	if !info.IsDir() {
		fmt.Printf("Invalid directory path %s\n", directory)
		os.Exit(1)
	}

	protocol := "tcp"
	address := fmt.Sprintf("%s:%d", host, port)

	listener, err := net.Listen(protocol, address)
	if err != nil {
		fmt.Printf("Failed to bind to port %d\n", port)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Printf("Listening for connections on %s\n", address)
	fmt.Printf("Serving files from %s\n", directory)

	for {
		connection, err := listener.Accept()
		if err != nil {
			fmt.Printf("Error accepting connection: %v\n", err)
		}
		fmt.Printf("Client connected %v\n", connection.RemoteAddr())
		go handleConnection(connection, directory)
	}

}
