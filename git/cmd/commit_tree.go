package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
)

func gitCommitTree(args []string) {
	usage := false
	if len(args) < 3 {
		usage = true
	}

	var treeHash, parentHash, message string
	for i := 0; !usage && i < len(args); i++ {
		switch args[i] {
		case "-p":
			if parentHash == "" && i+1 < len(args) {
				parentHash = args[i+1]
			} else {
				usage = true
			}
			i++ // skip
		case "-m":
			if message == "" && i+1 < len(args) {
				message = args[i+1]
			} else {
				usage = true
			}
			i++ // skip
		default:
			treeHash = args[i]
		}
	}

	if usage {
		printUsageAndExit("commit-tree <tree_sha> [-p <parent_sha>] [-m <message>] \n")
		return
	}

	// make sure tree_sha and parent_sha exists and have the right type
	treeType, _ := getObjTypeAndSize(treeHash)
	if treeType != "tree" {
		fatal("expected '%s' to be a 'tree' object, got: %s", treeHash, treeType)
	}

	parentType, _ := getObjTypeAndSize(treeHash)
	if parentType != "tree" {
		fatal("expected '%s' to be a 'tree' object, got: %s", parentHash, parentType)
	}

	username, email := getGitConfig("user.name"), getGitConfig("user.email")
	now := time.Now()
	timestamp := now.Unix()
	_, tzOffset := now.Zone()
	tzHours, tzMinutes := tzOffset/3600, (tzOffset/60)%60
	timezone := tzHours*100 + tzMinutes

	content := fmt.Sprintf("tree %s\n", treeHash)
	if parentHash != "" {
		content += fmt.Sprintf("parent %s\n", parentHash)
	}
	content += fmt.Sprintf("author %s <%s> %d %+05d\n", username, email, timestamp, timezone)
	content += fmt.Sprintf("committer %s <%s> %d %+05d\n", username, email, timestamp, timezone)
	content += fmt.Sprintf("\n%s\n", message)

	commitHash := hashObject(true, "commit", int64(len(content)), []byte(content))
	fmt.Printf("%x\n", commitHash)
}

func getGitConfig(key string) string {
	cmd := exec.Command("git", "config", key)
	output, err := cmd.Output()
	if err != nil {
		fatal(err.Error())
	}
	return strings.TrimRight(string(output), "\r\n")
}
