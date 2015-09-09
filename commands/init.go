package commands

import (
	"github.com/spf13/cobra"
	"os"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init files-tree",
	Long:  `init files-tree`,
	Run: func(cmd *cobra.Command, args []string){
		buildArgs.Path = ""
		if len(os.Args) > 2 {
			buildArgs.Path = os.Args[2] ;
		}
		discoverAndRunInitType(".", buildArgs)
	},
}

