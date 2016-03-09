package main

import (
	"bufio"
	"fmt"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/coreos/go-semver/semver"
	"github.com/n0rad/go-erlog"
	"github.com/n0rad/go-erlog/logs"
	_ "github.com/n0rad/go-erlog/register"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var CommitHash string
var DgrVersion string
var BuildDate string
var Args = BuildArgs{}
var workPath string

type BuildArgs struct {
	Force       bool
	Clean       bool
	Test        bool
	NoTestFail  bool
	KeepBuilder bool
}

func main() {
	logs.GetDefaultLog().(*erlog.ErlogLogger).Appenders[0].(*erlog.ErlogWriterAppender).Out = os.Stdout

	if os.Getuid() != 0 {
		println("dgr needs to be run as root")
		os.Exit(1)
	}

	Execute()
}

const RKT_SUPPORTED_VERSION = "0.12.0"

func Execute() {
	checkRktVersion()

	var version bool
	var homePath string
	var targetRootPath string
	var logLevel string

	var rootCmd = &cobra.Command{
		Use: "dgr",
		Run: func(cmd *cobra.Command, args []string) {},
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
	rootCmd.PersistentFlags().StringVarP(&homePath, "home-path", "H", DefaultHomeFolder(""), "Set home folder")
	rootCmd.PersistentFlags().StringVarP(&workPath, "work-path", "W", ".", "Set the work path")
	rootCmd.PersistentFlags().BoolVarP(&version, "version", "V", false, "Display dgr version")

	rootCmd.AddCommand(buildCmd, cleanCmd, pushCmd, installCmd, testCmd, versionCmd, initCmd, graphCmd, aciVersion)

	rootCmd.Execute()

	logs.Debug("Victory !")
}

func checkRktVersion() {
	output, err := common.ExecCmdGetOutput("rkt", "version")
	if err != nil {
		logs.Fatal("rkt is required in PATH")
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "VERSION:" {
			scanner.Scan()
			versionString := strings.TrimSpace(scanner.Text())
			version, err := semver.NewVersion(versionString)
			if err != nil {
				logs.WithField("content", version).Fatal("Cannot parse version of rkt")
			}
			supported, _ := semver.NewVersion(RKT_SUPPORTED_VERSION)
			if version.LessThan(*supported) {
				logs.WithField("requires", ">="+RKT_SUPPORTED_VERSION).Fatal("rkt version in your path is too old")
			}
			break
		}
	}
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
