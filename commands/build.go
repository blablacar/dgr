package commands

import "github.com/spf13/cobra"

var buildCmd = &cobra.Command{
	Use:   "build",
	Short: "build aci or pod",
	Long:  `build an aci or a pod`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunBuildType(".", buildArgs)
	},
}

func init() {
	buildCmd.Flags().BoolVarP(&buildArgs.ForceUpdate, "force-update", "U", true, "Force update of dependencies")
}
