package main

import (
	"github.com/n0rad/go-erlog/logs"        // the api
	_ "github.com/n0rad/go-erlog/register" // use erlog implementation, with default appender (colored to stderr)
)

func main() {
	logs.SetLevel(logs.TRACE) // default is INFO

	logs.Trace("I'm trace")
	logs.Debug("I'm debug")
	logs.Info("I'm info")
	logs.Warn("I'm warn")
	logs.Error("I'm error")

	func() {
		defer func() { recover() }()
		func() { logs.Panic("I'm panic") }()
	}()

	logs.Fatal("I'm fatal")
}
