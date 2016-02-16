package commands

import (
	"fmt"
	"github.com/blablacar/dgr/dgr"
	"github.com/blablacar/dgr/utils"
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

		runCleanIfRequested(workPath, buildArgs)
		buildAciOrPod(workPath, buildArgs).Build()
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean build",
	Long:  `clean build, including rootfs`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		buildAciOrPod(workPath, buildArgs).Clean()
	},
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "generate graphviz part",
	Long:  `generate graphviz part`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		buildAciOrPod(workPath, buildArgs).Graph()
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install image(s)",
	Long:  `install image(s) to rkt local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		runCleanIfRequested(workPath, buildArgs)
		buildAciOrPod(workPath, buildArgs).Install()
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push image(s)",
	Long:  `push images to repository`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		runCleanIfRequested(workPath, buildArgs)
		buildAciOrPod(workPath, buildArgs).Push()
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test image(s)",
	Long:  `test image(s)`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		runCleanIfRequested(workPath, buildArgs)
		buildAciOrPod(workPath, buildArgs).Test()
	},
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update aci",
	Long:  `update an aci`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		runCleanIfRequested(workPath, buildArgs)
		buildAciOrPod(workPath, buildArgs).Update()
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
		im := utils.ExtractManifestFromAci(args[0])
		val, _ := im.Labels.Get("version")
		println(val)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of dgr",
	Long:  `Print the version number of dgr`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		fmt.Print("dgr\n\n")
		fmt.Printf("version    : %s\n", dgr.Version)
		if dgr.BuildDate != "" {
			fmt.Printf("build date : %s\n", dgr.BuildDate)
		}
		if dgr.CommitHash != "" {
			fmt.Printf("CommitHash : %s\n", dgr.CommitHash)
		}
		os.Exit(0)
	},
}

func checkNoArgs(args []string) {
	if len(args) > 0 {
		logs.WithField("args", args).Fatal("Unknown arguments")
	}
}

func init() {
	installCmd.Flags().BoolVarP(&buildArgs.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
	installCmd.Flags().BoolVarP(&buildArgs.Test, "test", "t", false, "Run tests before install")

	pushCmd.Flags().BoolVarP(&buildArgs.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
	pushCmd.Flags().BoolVarP(&buildArgs.Test, "test", "t", false, "Run tests before push")

	initCmd.Flags().BoolVarP(&initForce, "force", "f", false, "Force init command if path is not empty")

	testCmd.Flags().BoolVarP(&buildArgs.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
}
