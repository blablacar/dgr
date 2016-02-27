package main

import (
	"bufio"
	"fmt"
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/coreos/go-semver/semver"
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
	if os.Getuid() != 0 {
		println("dgr needs to be run as root")
		os.Exit(1)
	}

	Execute()
}

const RKT_SUPPORTED_VERSION = "0.12.0"

func Execute() {
	checkRktVersion()

	var logLevel string
	var rootCmd = &cobra.Command{
		Use: "dgr",
	}
	var homePath string
	var targetRootPath string
	rootCmd.PersistentFlags().BoolVarP(&Args.Clean, "clean", "c", false, "Clean before doing anything")
	rootCmd.PersistentFlags().StringVarP(&targetRootPath, "targets-root-path", "p", "", "Set targets root path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "L", "info", "Set log level")
	rootCmd.PersistentFlags().StringVarP(&homePath, "home-path", "H", DefaultHomeFolder(""), "Set home folder")
	rootCmd.PersistentFlags().StringVarP(&workPath, "work-path", "W", ".", "Set the work path")

	rootCmd.AddCommand(buildCmd, cleanCmd, pushCmd, installCmd, testCmd, versionCmd, initCmd, graphCmd, aciVersion)
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {

		// logs

		level, err := logs.ParseLevel(logLevel)
		if err != nil {
			fmt.Printf("Unknown log level : %s", logLevel)
			os.Exit(1)
		}
		logs.SetLevel(level)

		Home = NewHome(homePath)

		// targetRootPath
		if targetRootPath != "" {
			Home.Config.TargetWorkDir = targetRootPath
		}

	}

	//
	//	if config.GetConfig().TargetWorkDir != "" {
	//		buildArgs.TargetsRootPath = config.GetConfig().TargetWorkDir
	//	}

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
				panic("Cannot parse version of rkt" + versionString)
			}
			supported, _ := semver.NewVersion(RKT_SUPPORTED_VERSION)
			if version.LessThan(*supported) {
				panic("rkt version in your path is too old. Require >= " + RKT_SUPPORTED_VERSION)
			}
			break
		}
	}

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
		NewAciOrPod(path, args).Clean()
	}
}
