package main

import (
	"github.com/blablacar/dgr/aci-builder/bin-run/builder"
	"github.com/n0rad/go-erlog"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"os"
)

func main() {
	logs.GetDefaultLog().(*erlog.ErlogLogger).Appenders[0].(*erlog.ErlogWriterAppender).Out = os.Stdout

	uuid := ProcessArgsAndReturnPodUUID()

	dir, err := os.Getwd()
	if err != nil {
		logs.WithE(err).Fatal("Failed to get current working directory")
	}

	b, err := builder.NewBuilder(dir, uuid)
	if err != nil {
		logs.WithE(err).Fatal("Failed to load Builder")
	}

	if err = b.Build(); err != nil {
		logs.WithE(err).Fatal("Build failed")
	}

	os.Exit(0)
}
