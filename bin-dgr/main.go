package main

import (
	"bufio"
	"fmt"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog"
	"github.com/n0rad/go-erlog/errs"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

const dgrEnvPrefix = "DGR_ENV_"

var CommitHash string
var DgrVersion string
var BuildDate string
var Args = BuildArgs{}
var workPath string

type envMap struct {
	mapping map[string]string
}

func (e *envMap) Set(s string) error {
	if e.mapping == nil {
		e.mapping = make(map[string]string)
	}
	pair := strings.SplitN(s, "=", 2)
	if len(pair) != 2 {
		return errs.With("environment variable must be specified as name=value")
	}
	e.mapping[pair[0]] = pair[1]
	return nil
}

func (e *envMap) String() string {
	return strings.Join(e.Strings(), "\n")
}

func (e *envMap) Strings() []string {
	var env []string
	for n, v := range e.mapping {
		env = append(env, n+"="+v)
	}
	return env
}

func (e *envMap) Type() string {
	return "envMap"
}

type BuildArgs struct {
	NoStore     bool
	StoreOnly   bool
	Force       bool
	Clean       bool
	Test        bool
	NoTestFail  bool
	KeepBuilder bool
	TrapOnError bool
	SetEnv      envMap
}

func main() {
	logs.GetDefaultLog().(*erlog.ErlogLogger).Appenders[0].(*erlog.ErlogWriterAppender).Out = os.Stdout

	if os.Getuid() != 0 {
		println("dgr needs to be run as root")
		os.Exit(1)
	}

	if !SupportsOverlay() {
		logs.Fatal("Overlay filesystem is required")
	}

	Execute()
}

func Execute() {
	var version bool
	var homePath string
	var targetRootPath string
	var logLevel string

	var rootCmd = &cobra.Command{
		Use: "dgr",
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
			os.Exit(1)
		},
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if version {
				displayVersionAndExit()
			}

			level, err := logs.ParseLevel(logLevel)
			if err != nil {
				fmt.Printf("Unknown log level : %s", logLevel)
				os.Exit(1)
			}
			logs.SetLevel(level)

			Home = NewHome(homePath)

			if targetRootPath != "" {
				Home.Config.TargetWorkDir = targetRootPath
			}
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&Args.Clean, "clean", "c", false, "Clean before doing anything")
	rootCmd.PersistentFlags().StringVarP(&targetRootPath, "targets-root-path", "p", "", "Set targets root path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "L", "info", "Set log level")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set log level")
	rootCmd.PersistentFlags().StringVarP(&homePath, "home-path", "H", DefaultHomeFolder(""), "Set home folder")
	rootCmd.PersistentFlags().StringVarP(&workPath, "work-path", "W", ".", "Set the work path")
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "V", false, "Display dgr version")
	rootCmd.PersistentFlags().Var(&Args.SetEnv, "set-env", "Env passed to builder scripts")
	rootCmd.PersistentFlags().BoolVar(&Args.StoreOnly, "store-only", false, "Tell rkt to use the store only")
	rootCmd.PersistentFlags().BoolVar(&Args.NoStore, "no-store", false, "Tell rkt to not use store")

	rootCmd.AddCommand(buildCmd, cleanCmd, pushCmd, installCmd, testCmd, versionCmd, initCmd, graphCmd, tryCmd /*signCmd,*/, aciVersion)

	readEnvironment()
	rootCmd.Execute()

	logs.Debug("Victory !")
}

func readEnvironment() {
	for _, v := range os.Environ() {
		if !strings.HasPrefix(v, dgrEnvPrefix) {
			continue
		}
		Args.SetEnv.Set(v[len(dgrEnvPrefix):])
	}
}

func SupportsOverlay() bool {
	exec.Command("modprobe", "overlay").Run()

	f, err := os.Open("/proc/filesystems")
	if err != nil {
		fmt.Println("error opening /proc/filesystems")
		return false
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		if s.Text() == "nodev\toverlay" {
			return true
		}
	}
	return false
}

func displayVersionAndExit() {
	fmt.Print("dgr\n\n")
	fmt.Printf("version    : %s\n", DgrVersion)
	if BuildDate != "" {
		fmt.Printf("build date : %s\n", BuildDate)
	}
	if CommitHash != "" {
		fmt.Printf("CommitHash : %s\n", CommitHash)
	}
	os.Exit(0)
}

func NewAciOrPod(path string, args BuildArgs) DgrCommand {
	if aci, err := NewAci(path, args); err == nil {
		return aci
	} else if pod, err2 := NewPod(path, args); err2 == nil {
		return pod
	} else {
		logs.WithField("path", path).WithField("err", err).WithField("err2", err2).Fatal("Cannot construct aci or pod")
	}
	return nil
}

func runCleanIfRequested(path string, args BuildArgs) {
	if args.Clean {
		logs.Warn("-c is deprecated and will be removed. Use 'dgr clean test', 'dgr clean install', 'dgr clean push' instead")
		NewAciOrPod(path, args).Clean()
	}
}

func giveBackUserRights(path string) {
	uid := "0"
	gid := "0"
	if os.Getenv("SUDO_UID") != "" {
		uid = os.Getenv("SUDO_UID")
		gid = os.Getenv("SUDO_GID")
	}
	common.ExecCmd("chown", "-R", uid+":"+gid, path)
}
