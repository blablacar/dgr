package commands

import (
	"bufio"
	"github.com/blablacar/cnt/builder"
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/logger"
	"github.com/blablacar/cnt/utils"
	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"strings"
)

var buildArgs = builder.BuildArgs{}

const RKT_SUPPORTED_VERSION = "0.8.1"

func Execute() {
	log.Logger = logger.NewLogger()
	checkRktVersion()

	var rootCmd = &cobra.Command{Use: "cnt"}
	rootCmd.PersistentFlags().BoolVarP(&buildArgs.Clean, "clean", "c", false, "Clean before doing anything")
	rootCmd.PersistentFlags().StringVarP(&buildArgs.TargetPath, "target-path", "p", "", "Set target path")

	rootCmd.AddCommand(buildCmd, cleanCmd, pushCmd, installCmd, testCmd, versionCmd, initCmd, updateCmd, graphCmd)

	config.GetConfig().Load()
	rootCmd.Execute()

	log.Info("Victory !")
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

func runCleanIfRequested(path string, args builder.BuildArgs) {
	if args.Clean {
		discoverAndRunCleanType(path, args)
	}
}
