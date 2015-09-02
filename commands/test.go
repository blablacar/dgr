package commands

import "github.com/spf13/cobra"

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test image(s)",
	Long:  `test image(s)`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunTestType(".", buildArgs)
	},
}
