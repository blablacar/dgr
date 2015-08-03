package commands

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"github.com/blablacar/cnt/application"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Version of cnt",
	Long:  `Print the version number of cnt`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("Cnt\n\n");
		fmt.Printf("version    : %s\n", application.Version)
		if application.BuildDate != "" {
			fmt.Printf("build date : %s\n", application.BuildDate)
		}
		if (application.CommitHash != "") {
			fmt.Printf("CommitHash : %s\n", application.CommitHash)
		}
		os.Exit(0)
	},
}
