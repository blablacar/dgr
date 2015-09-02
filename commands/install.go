package commands

import "github.com/spf13/cobra"

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install image(s)",
	Long:  `install image(s) to rkt local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunInstallType(".", buildArgs)
	},
}
