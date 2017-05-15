package main

import (
	"flag"
	"fmt"
	"net"
	"os"

	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/logs"
	rktcommon "github.com/rkt/rkt/common"
	stage1commontypes "github.com/rkt/rkt/stage1/common/types"
)

var cliDebugFlag bool

var (
	debug       bool
	localhostIP net.IP
	localConfig string
)

func parseFlags() *stage1commontypes.RuntimePod {
	rp := stage1commontypes.RuntimePod{}

	flag.BoolVar(&debug, "debug", false, "Run in debug mode")
	flag.StringVar(&localConfig, "local-config", rktcommon.DefaultLocalConfigDir, "Local config path")

	// These flags are persisted in the PodRuntime
	flag.BoolVar(&rp.Interactive, "interactive", false, "The pod is interactive")
	flag.Var(&rp.NetList, "net", "Setup networking")
	flag.StringVar(&rp.MDSToken, "mds-token", "", "MDS auth token")

	flag.Parse()

	rp.Debug = debug

	return &rp
}

func ProcessArgsAndReturnPodUUID() (*types.UUID, *stage1commontypes.RuntimePod) {
	rp := parseFlags()

	if debug {
		logs.SetLevel(logs.DEBUG)
	}
	if lvlStr := os.Getenv(common.EnvLogLevel); lvlStr != "" {
		lvl, err := logs.ParseLevel(lvlStr)
		if err != nil {
			fmt.Printf("Unknown log level : %s", lvlStr)
			os.Exit(1)
		}
		logs.SetLevel(lvl)
	}

	arg := flag.Arg(0)
	uuid, err := types.NewUUID(arg)
	if err != nil {
		logs.WithE(err).WithField("content", arg).Fatal("UUID is missing or malformed")
	}
	return uuid, rp
}
