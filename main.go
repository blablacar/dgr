package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/commands"
	"github.com/blablacar/cnt/log"
	"os"
)

//go:generate go run compile/info_generate.go
func main() {
	logrus.SetFormatter(&log.BlaFormatter{})

	if os.Getuid() != 0 {
		println("Cnt needs to be run as root")
		os.Exit(1)
	}

	commands.Execute()
}
