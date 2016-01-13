package commands

import (
	"bufio"
	"fmt"
	"github.com/blablacar/cnt/builder"
	"github.com/blablacar/cnt/cnt"
	"github.com/blablacar/cnt/spec"
	"github.com/blablacar/cnt/utils"
	"github.com/coreos/go-semver/semver"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var buildArgs = builder.BuildArgs{}

const RKT_SUPPORTED_VERSION = "0.12.0"

var workPath string

func Execute() {
	checkRktVersion()

	var logLevel string
	var rootCmd = &cobra.Command{
		Use: "cnt",
	}
	var homePath string
	var targetRootPath string
	rootCmd.PersistentFlags().BoolVarP(&buildArgs.Clean, "clean", "c", false, "Clean before doing anything")
	rootCmd.PersistentFlags().StringVarP(&targetRootPath, "targets-root-path", "p", "", "Set targets root path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "L", "info", "Set log level")
	rootCmd.PersistentFlags().StringVarP(&homePath, "home-path", "H", cnt.DefaultHomeFolder(), "Set home folder")
	rootCmd.PersistentFlags().StringVarP(&workPath, "work-path", "W", ".", "Set the work path")

	rootCmd.AddCommand(buildCmd, cleanCmd, pushCmd, installCmd, testCmd, versionCmd, initCmd /*updateCmd,*/, graphCmd, aciVersion)
	rootCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {

		// logs

		level, err := logs.ParseLevel(logLevel)
		if err != nil {
			fmt.Printf("Unknown log level : %s", logLevel)
			os.Exit(1)
		}
		logs.SetLevel(level)

		cnt.Home = cnt.NewHome(homePath)

		// targetRootPath
		if targetRootPath != "" {
			cnt.Home.Config.TargetWorkDir = targetRootPath
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
	output, err := utils.ExecCmdGetOutput("rkt")
	if err != nil {
		panic("rkt is required in PATH")
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

func buildAciOrPod(path string, args builder.BuildArgs) spec.CntCommand {
	if aci, err := builder.NewAci(path, args); err == nil {
		return aci
	} else if pod, err2 := builder.NewPod(path, args); err2 == nil {
		return pod
	} else {
		logs.WithField("path", path).WithField("err", err).WithField("err2", err2).Fatal("Cannot construct aci or pod")
	}
	return nil
}

func runCleanIfRequested(path string, args builder.BuildArgs) {
	if args.Clean {
		buildAciOrPod(path, args).Clean()
	}
}
