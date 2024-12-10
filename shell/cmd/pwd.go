package main

import (
	"fmt"
	"os"
)

func pwd(_ []string) {
	dirPath, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(dirPath)
}
