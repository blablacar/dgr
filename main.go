package main

import (
	"github.com/blablacar/dgr/commands"
	_ "github.com/n0rad/go-erlog/register"
	"os"
)

//go:generate go run compile/info_generate.go
func main() {

	if os.Getuid() != 0 {
		println("dgr needs to be run as root")
		os.Exit(1)
	}

	commands.Execute()
}
