package main

import (
	"github.com/blablacar/cnt/commands"
	"os"
)

//go:generate go run compile/info_generate.go
func main() {
	if os.Getuid() != 0 {
		println("Cnt needs to be run as root")
		os.Exit(1)
	}

	commands.Execute()
}
