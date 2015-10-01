package commands

import (
	"github.com/blablacar/cnt/builder"
	"github.com/blablacar/cnt/log"
	"github.com/spf13/cobra"
)

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push image(s)",
	Long:  `push images to repository`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunPushType(".", buildArgs)
	},
}

func discoverAndRunPushType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Push()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Push()
	} else {
		log.Get().Panic("Cannot find cnt-manifest.yml")
	}
}
