package main

import (
	"os"
	"github.com/spf13/cobra"
	"github.com/blablacar/cnt/builder"
	"github.com/blablacar/cnt/config"
	"github.com/blablacar/cnt/log"
	"github.com/blablacar/cnt/logger"
)

func main() {
	if os.Getuid() != 0 {
		log.Get().Panic("Please run this command as root")
	}

	processArgs();
}

func discoverAndRunBuildType(path string, args builder.BuildArgs) {
 	if cnt, err := builder.OpenCnt(path, args); err == nil {
		cnt.Build()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Build()
	} else {
		log.Get().Panic("Cannot found cnt-manifest.yml")
	}
}

func discoverAndRunPushType(path string, args builder.BuildArgs) {
	if cnt, err := builder.OpenCnt(path, args); err == nil {
		cnt.Push()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Push()
	} else {
		log.Get().Panic("Victory Cannot found cnt-manifest.yml")
	}
}

func discoverAndRunInstallType(path string, args builder.BuildArgs) {
	if cnt, err := builder.OpenCnt(path, args); err == nil {
		cnt.Install()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Install()
	} else {
		log.Get().Panic("Victory Cannot found cnt-manifest.yml")
	}
}

func processArgs() {
	log.Set(logger.NewLogger())

	buildArgs := builder.BuildArgs{}

	var cmdBuild = &cobra.Command{
		Use:   "build",
		Short: "build aci or pod",
		Long:  `build an aci or a pod`,
		Run: func(cmd *cobra.Command, args []string) {
			discoverAndRunBuildType(".", buildArgs)
		},
	}
	cmdBuild.Flags().BoolVarP(&buildArgs.Zip, "nozip", "z", true, "zip final image or not")

	var cmdClean = &cobra.Command{
		Use:   "clean",
		Short: "clean build",
		Long:  `clean build, including rootfs`,
		Run: func(cmd *cobra.Command, args []string) {
			os.RemoveAll("target/");
		},
	}

	var push = &cobra.Command{
		Use:   "push",
		Short: "push image(s)",
		Long:  `push images to repository`,
		Run: func(cmd *cobra.Command, args []string) {
			discoverAndRunPushType(".", buildArgs)
		},
	}

	var install = &cobra.Command{
		Use:   "install",
		Short: "install image(s)",
		Long:  `install image(s) to rkt local storage`,
		Run: func(cmd *cobra.Command, args []string) {
			discoverAndRunInstallType(".", buildArgs)
		},
	}

	var rootCmd = &cobra.Command{Use: "cnt"}
	rootCmd.AddCommand(cmdBuild, cmdClean, push, install)

	config.GetConfig().Load()
	rootCmd.Execute()

	log.Get().Info("Victory !")
}

