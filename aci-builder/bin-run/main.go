package main

import (
	"os"

	"github.com/n0rad/go-erlog"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
)

func main() {
	logs.GetDefaultLog().(*erlog.ErlogLogger).Appenders[0].(*erlog.ErlogWriterAppender).Out = os.Stdout

	uuid, rp := ProcessArgsAndReturnPodUUID()

	dir, err := os.Getwd()
	if err != nil {
		logs.WithE(err).Fatal("Failed to get current working directory")
	}

	b, err := NewBuilder(dir, uuid, rp)
	if err != nil {
		logs.WithE(err).Fatal("Failed to load Builder")
	}

	if err = b.Build(); err != nil {
		logs.WithE(err).Fatal("Build failed")
	}

	os.Exit(0)
}
