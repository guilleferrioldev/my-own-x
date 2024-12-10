package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

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
