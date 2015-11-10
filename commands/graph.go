package commands

import (
	"github.com/blablacar/cnt/builder"
	"github.com/spf13/cobra"
	"os"
)

var graphCmd = &cobra.Command{
	Use:   "graph",
	Short: "generate graphviz part",
	Long:  `generate graphviz part`,
	Run: func(cmd *cobra.Command, args []string) {
		buildArgs.Path = ""
		if len(os.Args) > 2 {
			buildArgs.Path = os.Args[2]
		}
		discoverAndRunGraphType(".", buildArgs)
	},
}

func discoverAndRunGraphType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Graph()
	} else if pod, err2 := builder.NewPod(path, args); err2 == nil {
		pod.Graph()
	} else {
		panic("Cannot find cnt-manifest.yml or cnt-pod-manifest.yml" + err.Error() + err2.Error())
	}
}
