package main

import (
	"fmt"
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
		buildAciOrPod(workPath, Args).Build()
	},
}

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean build",
	Long:  `clean build, including rootfs`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		buildAciOrPod(workPath, Args).Clean()
	},
}

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "generate graphviz part",
	Long:  `generate graphviz part`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		buildAciOrPod(workPath, Args).Graph()
	},
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install image(s)",
	Long:  `install image(s) to rkt local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		runCleanIfRequested(workPath, Args)
		buildAciOrPod(workPath, Args).Install()
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push image(s)",
	Long:  `push images to repository`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		runCleanIfRequested(workPath, Args)
		buildAciOrPod(workPath, Args).Push()
	},
}

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test image(s)",
	Long:  `test image(s)`,
	Run: func(cmd *cobra.Command, args []string) {
		checkNoArgs(args)

		runCleanIfRequested(workPath, Args)
		buildAciOrPod(workPath, Args).Test()
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
		im := ExtractManifestFromAci(args[0])
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
		fmt.Printf("version    : %s\n", DgrVersion)
		if BuildDate != "" {
			fmt.Printf("build date : %s\n", BuildDate)
		}
		if CommitHash != "" {
			fmt.Printf("CommitHash : %s\n", CommitHash)
		}
		os.Exit(0)
	},
}

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init files-tree",
	Long:  `init files-tree`,
	Run: func(cmd *cobra.Command, args []string) {
		//TODO
		uid := "0"
		gid := "0"
		if os.Getenv("SUDO_UID") != "" {
			uid = os.Getenv("SUDO_UID")
			gid = os.Getenv("SUDO_GID")
		}
		common.ExecCmd("chown", "-R", uid+":"+gid, workPath)
	},
}

func checkNoArgs(args []string) {
	if len(args) > 0 {
		logs.WithField("args", args).Fatal("Unknown arguments")
	}
}

func init() {
	buildCmd.Flags().BoolVarP(&Args.KeepBuilder, "keep-builder", "k", false, "Keep builder container after exit")

	installCmd.Flags().BoolVarP(&Args.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
	installCmd.Flags().BoolVarP(&Args.Test, "test", "t", false, "Run tests before install")

	pushCmd.Flags().BoolVarP(&Args.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
	pushCmd.Flags().BoolVarP(&Args.Test, "test", "t", false, "Run tests before push")

	initCmd.Flags().BoolVarP(&Args.Force, "force", "f", false, "Force init command if path is not empty")

	testCmd.Flags().BoolVarP(&Args.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
}
