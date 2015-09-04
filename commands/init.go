package commands

import "github.com/spf13/cobra"

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "init files-tree",
	Long:  `init files-tree`,
	Run: func(cmd *cobra.Command, args []string) {
		discoverAndRunInitType(".", buildArgs)
	},
}

func init(){
	initCmd.Flags().StringVarP(&buildArgs.Path, "path", "p", "", "Specify a path : -p /my/path")
}
