package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"runtime"

	"github.com/appc/spec/schema/types"
	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog/logs"
	rktcommon "github.com/rkt/rkt/common"
	pkgflag "github.com/rkt/rkt/pkg/flag"
	rktlog "github.com/rkt/rkt/pkg/log"
	stage1commontypes "github.com/rkt/rkt/stage1/common/types"
)

var cliDebugFlag bool

var (
	debug       bool
	localhostIP net.IP
	localConfig string
	log         *rktlog.Logger
	diag        *rktlog.Logger
	interpBin   string // Path to the interpreter within the stage1 rootfs, set by the linker
)

func parseFlags() *stage1commontypes.RuntimePod {
	rp := stage1commontypes.RuntimePod{}

	flag.BoolVar(&debug, "debug", false, "Run in debug mode")
	flag.StringVar(&localConfig, "local-config", rktcommon.DefaultLocalConfigDir, "Local config path")

	// These flags are persisted in the PodRuntime
	flag.BoolVar(&rp.Interactive, "interactive", false, "The pod is interactive")
	flag.BoolVar(&rp.Mutable, "mutable", false, "Enable mutable operations on this pod, including starting an empty one")
	flag.Var(&rp.NetList, "net", "Setup networking")
	flag.StringVar(&rp.PrivateUsers, "private-users", "", "Run within user namespace. Can be set to [=UIDBASE[:NUIDS]]")
	flag.StringVar(&rp.MDSToken, "mds-token", "", "MDS auth token")
	flag.StringVar(&rp.Hostname, "hostname", "", "Hostname of the pod")
	flag.BoolVar(&rp.InsecureOptions.DisableCapabilities, "disable-capabilities-restriction", false, "Disable capability restrictions")
	flag.BoolVar(&rp.InsecureOptions.DisablePaths, "disable-paths", false, "Disable paths restrictions")
	flag.BoolVar(&rp.InsecureOptions.DisableSeccomp, "disable-seccomp", false, "Disable seccomp restrictions")
	dnsConfMode := pkgflag.MustNewPairList(map[string][]string{
		"resolv": {"host", "stage0", "none", "default"},
		"hosts":  {"host", "stage0", "default"},
	}, map[string]string{
		"resolv": "default",
		"hosts":  "default",
	})
	flag.Var(dnsConfMode, "dns-conf-mode", "DNS config file modes")

	flag.Parse()

	rp.Debug = debug
	rp.ResolvConfMode = dnsConfMode.Pairs["resolv"]
	rp.EtcHostsMode = dnsConfMode.Pairs["hosts"]

	return &rp
}

func init() {
	// this ensures that main runs only on main thread (thread group leader).
	// since namespace ops (unshare, setns) are done for a single thread, we
	// must ensure that the goroutine does not jump from OS thread to thread
	runtime.LockOSThread()

	// We'll need this later
	localhostIP = net.ParseIP("127.0.0.1")
	if localhostIP == nil {
		panic("localhost IP failed to parse")
	}
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
