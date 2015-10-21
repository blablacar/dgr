package commands

import (
	"github.com/blablacar/cnt/builder"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "clean build",
	Long:  `clean build, including rootfs`,
	Run: func(cmd *cobra.Command, args []string) {
		discoverAndRunCleanType(".", buildArgs)
	},
}

func discoverAndRunCleanType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Clean()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Clean()
	} else {
		panic("Cannot find cnt-manifest.yml")
	}
}
