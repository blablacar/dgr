package main

import (
	"github.com/n0rad/go-erlog/log"        // the api
	_ "github.com/n0rad/go-erlog/register" // use erlog implementation, with default appender (colored to stderr)
)

func main() {
	log.SetLevel(log.TRACE) // default is INFO

	log.Trace("I'm trace")
	log.Debug("I'm debug")
	log.Info("I'm info")
	log.Warn("I'm warn")
	log.Error("I'm error")

	func() {
		defer func() { recover() }()
		func() { log.Panic("I'm panic") }()
	}()

	log.Fatal("I'm fatal")
}
