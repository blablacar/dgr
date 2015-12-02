package commands

import (
	"fmt"
	"github.com/blablacar/cnt/cnt"
	"github.com/spf13/cobra"
	"os"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of cnt",
	Long:  `Print the version number of cnt`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Cnt\n\n")
		fmt.Printf("version    : %s\n", cnt.Version)
		if cnt.BuildDate != "" {
			fmt.Printf("build date : %s\n", cnt.BuildDate)
		}
		if cnt.CommitHash != "" {
			fmt.Printf("CommitHash : %s\n", cnt.CommitHash)
		}
		os.Exit(0)
	},
}
