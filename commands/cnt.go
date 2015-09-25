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
	log.Set(logger.NewLogger())
	checkRktVersion()

	var rootCmd = &cobra.Command{Use: "cnt"}
	rootCmd.PersistentFlags().BoolVarP(&buildArgs.Clean, "clean", "c", false, "Clean before doing anything")
	rootCmd.PersistentFlags().StringVarP(&buildArgs.TargetPath, "target-path", "t", "", "Set target path")

	rootCmd.AddCommand(buildCmd, cleanCmd, pushCmd, installCmd, testCmd, versionCmd, initCmd, updateCmd)

	config.GetConfig().Load()
	rootCmd.Execute()

	log.Get().Info("Victory !")
}

func checkRktVersion() {
	output, err := utils.ExecCmdGetOutput("rkt")
	if err != nil {
		log.Get().Panic("rkt is required in PATH")
	}

	scanner := bufio.NewScanner(strings.NewReader(output))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "VERSION:" {
			scanner.Scan()
			versionString := strings.TrimSpace(scanner.Text())
			version, err := semver.NewVersion(versionString)
			if err != nil {
				log.Get().Panic("Cannot parse version of rkt", versionString)
			}
			supported, _ := semver.NewVersion(RKT_SUPPORTED_VERSION)
			if version.LessThan(*supported) {
				log.Get().Panic("rkt version in your path is too old. Require >= " + RKT_SUPPORTED_VERSION)
			}
			break
		}
	}

}

func discoverAndRunBuildType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Build()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Build()
	} else {
		log.Get().Panic("Cannot find cnt-manifest.yml")
	}
}

func discoverAndRunPushType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Push()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Push()
	} else {
		log.Get().Panic("Cannot find cnt-manifest.yml")
	}
}

func discoverAndRunInstallType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Install()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Install()
	} else {
		log.Get().Panic("Cannot find cnt-manifest.yml")
	}
}

func discoverAndRunUpdateType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.UpdateConf()
	} else if _, err := builder.OpenPod(path, args); err == nil {
		log.Get().Panic("Not Yet implemented for pods")
	} else {
		log.Get().Panic("Cannot find cnt-manifest.yml")
	}
}

func discoverAndRunCleanType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Clean()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Clean()
	} else {
		log.Get().Panic("Cannot find cnt-manifest.yml")
	}
}

func discoverAndRunTestType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Test()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Test()
	} else {
		log.Get().Panic("Cannot find cnt-manifest.yml")
	}
}

func runCleanIfRequested(path string, args builder.BuildArgs) {
	if args.Clean {
		discoverAndRunCleanType(path, args)
	}
}

func discoverAndRunInitType(path string, args builder.BuildArgs) {
	if cnt, err := builder.PrepAci(path, args); err == nil {
		cnt.Init()
	}
}
