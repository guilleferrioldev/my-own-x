package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	cache := NewCache()

	for {
		if cache.queue.length == 0 {
			fmt.Println("Write what you want to cache")
		} else {
			fmt.Println("Write more things")
		}
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

		cache.Check(input)
		cache.Display()
	}
}
