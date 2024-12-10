package main

import (
	"fmt"
	"os"
	"strconv"
)

func exit(args []string) {
	if len(args) != 1 {
		fmt.Println("Invalid number of arguements for command exit. Expected 1, Got", len(args))
		return
	}

	exitCode, err := strconv.Atoi(args[0])
	if err != nil {
		fmt.Println("Invalid exit code " + args[0])
	}

	os.Exit(exitCode)
}
