package commands

import (
	"bufio"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/blablacar/cnt/builder"
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/utils"
	"github.com/coreos/go-semver/semver"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

var buildArgs = builder.BuildArgs{}

const RKT_SUPPORTED_VERSION = "0.11.0"

func Execute() {
	checkRktVersion()

	var logLevel string
	var rootCmd = &cobra.Command{
		Use: "cnt",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			level, err := log.ParseLevel(logLevel)
			if err != nil {
				fmt.Printf("Unknown log level : %s", logLevel)
				os.Exit(1)
			}
			log.SetLevel(level)
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&buildArgs.Clean, "clean", "c", false, "Clean before doing anything")
	rootCmd.PersistentFlags().StringVarP(&buildArgs.TargetsRootPath, "targets-root-path", "p", "", "Set targets root path")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "loglevel", "L", "info", "Set log level")

	rootCmd.AddCommand(buildCmd, cleanCmd, pushCmd, installCmd, testCmd, versionCmd, initCmd, updateCmd, graphCmd, aciVersion)

	config.GetConfig().Load()

	if config.GetConfig().TargetWorkDir != "" {
		buildArgs.TargetsRootPath = config.GetConfig().TargetWorkDir
	}

	rootCmd.Execute()

	log.Debug("Victory !")
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
