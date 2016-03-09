package main

import (
	"github.com/blablacar/dgr/bin-dgr/common"
	"github.com/n0rad/go-erlog/logs"
	"github.com/spf13/cobra"
	"os"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build aci or pod",
	Long:  `build an aci or a pod`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		runCleanIfRequested(workPath, Args)
		NewAciOrPod(workPath, Args).Build()
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean build",
	Long:  `clean build, including rootfs`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		NewAciOrPod(workPath, Args).Clean()
	},
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "generate graphviz part",
	Long:  `generate graphviz part`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		NewAciOrPod(workPath, Args).Graph()
	},
}

var aciVersion = &cobra.Command{
	Use:   "aci-version file",
	Short: "display version of aci",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			cmd.Usage()
			os.Exit(1)
		}
		im, err := common.ExtractManifestFromAci(args[0])
		if err != nil {
			logs.WithE(err).Fatal("Failed to get manifest from file")
		}
		val, _ := im.Labels.Get("version")
		println(val)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of dgr",
	Long:  `Print the version number of dgr`,
	Run: func(cmd *cobra.Command, args []string) {
		displayVersionAndExit()
	},
}

var installCmd = newInstallCommand(false)
var pushCmd = newPushCommand(false)
var testCmd = newTestCommand(false)

func checkNoArgs(args []string) {
	if len(args) > 0 {
		logs.WithField("args", args).Fatal("Unknown arguments")
	}
}

func newInstallCommand(underClean bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "install image(s)",
		Long:  `install image(s) to rkt local storage`,
		Run: func(cmd *cobra.Command, args []string) {
			checkNoArgs(args)

			command := NewAciOrPod(workPath, Args)
			if underClean {
				command.Clean()
			} else {
				runCleanIfRequested(workPath, Args)
			}
			command.Install()
		},
	}

	cmd.Flags().BoolVarP(&Args.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
	cmd.Flags().BoolVarP(&Args.Test, "test", "t", false, "Run tests before install")
	return cmd
}

func newPushCommand(underClean bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "push",
		Short: "push image(s)",
		Long:  `push images to repository`,
		Run: func(cmd *cobra.Command, args []string) {
			checkNoArgs(args)

			command := NewAciOrPod(workPath, Args)
			if underClean {
				command.Clean()
			} else {
				runCleanIfRequested(workPath, Args)
			}
			command.Push()
		},
	}
	cmd.Flags().BoolVarP(&Args.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
	cmd.Flags().BoolVarP(&Args.Test, "test", "t", false, "Run tests before push")
	return cmd
}

func newTestCommand(underClean bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "test image(s)",
		Long:  `test image(s)`,
		Run: func(cmd *cobra.Command, args []string) {
			checkNoArgs(args)

			command := NewAciOrPod(workPath, Args)
			if underClean {
				command.Clean()
			} else {
				runCleanIfRequested(workPath, Args)
			}
			command.Test()
		},
	}
	cmd.Flags().BoolVarP(&Args.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
	return cmd
}

func init() {
	cleanCmd.AddCommand(newInstallCommand(true))
	cleanCmd.AddCommand(newPushCommand(true))
	cleanCmd.AddCommand(newTestCommand(true))

	buildCmd.Flags().BoolVarP(&Args.KeepBuilder, "keep-builder", "k", false, "Keep builder container after exit")
}
