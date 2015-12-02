package commands

import (
	"github.com/blablacar/cnt/utils"
	"github.com/spf13/cobra"
	"os"
)

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
