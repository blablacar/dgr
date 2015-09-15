package commands

import "github.com/spf13/cobra"

var updateCmd = &cobra.Command{
Use:   "update",
Short: "update aci",
Long:  `update an aci`,
Run: func(cmd *cobra.Command, args []string) {
	runCleanIfRequested(".", buildArgs)
	discoverAndRunUpdateType(".", buildArgs)
},
}
