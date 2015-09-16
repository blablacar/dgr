package commands

import (
	"github.com/blablacar/cnt/builder"
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/logger"
	"github.com/spf13/cobra"
)

var buildArgs = builder.BuildArgs{}

func Execute() {
	log.Set(logger.NewLogger())
	var rootCmd = &cobra.Command{Use: "cnt"}
	buildCmd.Flags().BoolVarP(&buildArgs.Zip, "nozip", "z", false, "Zip final image or not")
	rootCmd.PersistentFlags().BoolVarP(&buildArgs.Clean, "clean", "c", false, "Clean before doing anything")
	rootCmd.PersistentFlags().StringVarP(&buildArgs.TargetPath, "target-path", "t", "", "Set target path")

	rootCmd.AddCommand(buildCmd, cleanCmd, pushCmd, installCmd, testCmd, versionCmd, initCmd, updateCmd)

	config.GetConfig().Load()
	rootCmd.Execute()

	log.Get().Info("Victory !")
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
