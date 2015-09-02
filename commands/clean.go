package commands

import "github.com/spf13/cobra"

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean build",
	Long:  `clean build, including rootfs`,
	Run: func(cmd *cobra.Command, args []string) {
		discoverAndRunCleanType(".", buildArgs)
	},
}
