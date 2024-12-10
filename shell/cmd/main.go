package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"unicode"
)

type Command func([]string)

var builtins = make(map[string]Command)

func main() {
	builtins = map[string]Command{
		"echo": echo,
		"exit": exit,
		"type": typeCommand,
		"pwd":  pwd,
		"cd":   cd,
	}

	for {
		fmt.Fprint(os.Stdout, "$ ")

		commandRaw, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			fmt.Println("Error In User Input")
		}

		command := parseRawCommand(commandRaw)
		if commandHandler, exists := builtins[command[0]]; exists {
			commandHandler(command[1:])
		} else if path, err := findExecutablePath(command[0]); err == nil {
			cmd := exec.Command(path, command[1:]...)
			cmd.Env = os.Environ()
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			err := cmd.Run()
			if err != nil {
				fmt.Println(err)
			}
		} else {
			fmt.Println(command[0] + ": command not found")
		}
	}
}

func parseRawCommand(command string) []string {
	var args []string
	var arg strings.Builder
	inSingleQuotes := false
	inDoubleQuotes := false
	escaped := false

	for _, char := range command {
		switch {
		case char == '\'':
			if escaped && inDoubleQuotes {
				arg.WriteRune('\\')
			}
			if inDoubleQuotes || escaped {
				arg.WriteRune(char)
			} else {
				inSingleQuotes = !inSingleQuotes
			}
			escaped = false
		case char == '"':
			if inSingleQuotes || escaped {
				arg.WriteRune(char)
			} else {
				inDoubleQuotes = !inDoubleQuotes
			}
			escaped = false
		case char == '\\':
			if inSingleQuotes || escaped {
				arg.WriteRune(char)
				escaped = false
			} else {
				escaped = true
			}
		case unicode.IsSpace(char):
			if escaped && (inDoubleQuotes || inSingleQuotes) {
				arg.WriteRune('\\')
			}
			if inSingleQuotes || inDoubleQuotes || escaped {
				arg.WriteRune(char)
			} else if arg.Len() > 0 {
				args = append(args, arg.String())
				arg.Reset()
			}
			escaped = false
		default:
			if escaped && inDoubleQuotes {
				arg.WriteRune('\\')
			}
			arg.WriteRune(char)
			escaped = false
		}
	}

	if arg.Len() > 0 {
		args = append(args, arg.String())
	}

	return args
}

func findExecutablePath(target string) (string, error) {
	pathEnv, isSet := os.LookupEnv("PATH")
	if !isSet {
		return "", errors.New("PATH variable is not set")
	}

	paths := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, dir := range paths {
		fullPath := filepath.Join(dir, target)
		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, nil
		}
	}

	return "", errors.New("executable file not found in PATH")
}
