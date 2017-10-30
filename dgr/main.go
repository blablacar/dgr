package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/blablacar/dgr/dgr/common"
	"github.com/n0rad/go-erlog"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/spf13/cobra"
)

const dgrEnvPrefix = "DGR_ENV_"

var BuildCommit string
var BuildVersion string
var BuildTime string

var Args = BuildArgs{}
var workPath string

type BuildArgs struct {
	Force         bool
	Test          bool
	NoTestFail    bool
	KeepBuilder   bool
	CatchOnError  bool
	CatchOnStep   bool
	ParallelBuild bool
	PullPolicy    string
	SetEnv        envMap
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
				fmt.Printf("Unknown log level : %s\n", logLevel)
				os.Exit(1)
			}
			logs.SetLevel(level)

			Home = NewHome(homePath)

			if targetRootPath != "" {
				Home.Config.TargetWorkDir = targetRootPath
			}

			if Args.PullPolicy != "" && !common.PullPolicy(Args.PullPolicy).IsValid() {
				logs.WithField("pull-policy", Args.PullPolicy).Fatal("Invalid pull-policy")
			}
		},
	}
	rootCmd.PersistentFlags().StringVarP(&targetRootPath, "targets-root-path", "p", "", "Set targets root path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "L", "info", "Set log level")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log-level", "info", "Set log level")
	rootCmd.PersistentFlags().StringVarP(&homePath, "home-path", "H", DefaultHomeFolder(""), "Set home folder")
	rootCmd.PersistentFlags().StringVarP(&workPath, "work-path", "W", ".", "Set the work path")
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "V", false, "Display dgr version")
	rootCmd.PersistentFlags().Var(&Args.SetEnv, "set-env", "Env passed to builder scripts")
	rootCmd.PersistentFlags().StringVar(&Args.PullPolicy, "pull-policy", "", "force rkt fetch Policy")
	rootCmd.PersistentFlags().BoolVarP(&Args.ParallelBuild, "parallel", "P", false, "Run build in parallel for pod")

	rootCmd.AddCommand(
		buildCmd,
		cleanCmd,
		pushCmd,
		installCmd,
		testCmd,
		versionCmd,
		initCmd,
		graphCmd,
		tryCmd,
		signCmd,
		aciVersion,
		configCmd,
		updateCmd,
		cpCmd,
	)

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
	fmt.Printf("version    : %s\n", BuildVersion)
	if BuildTime != "" {
		fmt.Printf("build date : %s\n", BuildTime)
	}
	if BuildCommit != "" {
		fmt.Printf("CommitHash : %s\n", BuildCommit)
	}
	os.Exit(0)
}

func NewAciOrPod(path string, args BuildArgs) DgrCommand {
	if aci, err := NewAci(path, args); err == nil {
		return aci
	} else if pod, err2 := NewPod(path, args); err2 == nil {
		return pod
	} else {
		logs.WithE(err).WithField("path", path).WithField("err2", err2).Fatal("Cannot construct aci or pod")
	}
	return nil
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
