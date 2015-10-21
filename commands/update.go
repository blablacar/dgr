package commands

import (
	"github.com/blablacar/cnt/builder"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "update aci",
	Long:  `update an aci`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunUpdateType(".", buildArgs)
	},
}

func discoverAndRunUpdateType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.UpdateConf()
	} else if _, err := builder.OpenPod(path, args); err == nil {
		panic("Not Yet implemented for pods")
	} else {
		panic("Cannot find cnt-manifest.yml")
	}
}
