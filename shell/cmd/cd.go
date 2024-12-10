package main

import (
	"fmt"
	"os"
)

func cd(args []string) {
	if len(args) != 1 {
		fmt.Println("Invalid number of arguements for command cd. Expected 1, Got", len(args))
		return
	}

	var targetDir string
	if args[0] == "~" {
		homeDir, isSet := os.LookupEnv("HOME")
		if !isSet {
			fmt.Println("HOME variable is not set")
			return
		}
		targetDir = homeDir
	} else {
		targetDir = args[0]
	}

	err := os.Chdir(targetDir)
	if err != nil {
		fmt.Println("cd:", targetDir+": No such file or directory")
		return
	}
}
