package commands

import (
	"github.com/blablacar/cnt/builder"
	"github.com/blablacar/cnt/log"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "install image(s)",
	Long:  `install image(s) to rkt local storage`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunInstallType(".", buildArgs)
	},
}

func discoverAndRunInstallType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Install()
	} else if pod, err := builder.OpenPod(path, args); err == nil {
		pod.Install()
	} else {
		log.Get().Panic("Cannot find cnt-manifest.yml")
	}
}
