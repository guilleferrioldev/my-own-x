package main

import (
	"fmt"
)

func typeCommand(args []string) {
	if len(args) != 1 {
		fmt.Println("Invalid number of arguements for command type. Expected 1, Got", len(args))
		return
	}

	if _, exists := builtins[args[0]]; exists {
		fmt.Println(args[0] + " is a shell builtin")
	} else if path, err := findExecutablePath(args[0]); err == nil {
		fmt.Println(args[0], "is", path)
	} else {
		fmt.Println(args[0] + ": not found")
	}
}
