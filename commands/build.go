package commands

import (
	"github.com/blablacar/cnt/builder"
	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build aci or pod",
	Long:  `build an aci or a pod`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunBuildType(".", buildArgs)
	},
}

func discoverAndRunBuildType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Build()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Build()
	} else {
		panic("Cannot find cnt-manifest.yml")
	}
}

func init() {
	//	buildCmd.Flags().BoolVarP(&buildArgs.ForceUpdate, "force-update", "U", false, "Force update of dependencies")
}
