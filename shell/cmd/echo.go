package main

import (
	"fmt"
	"strings"
)

func echo(args []string) {
	fmt.Println(strings.Join(args, " "))
}
