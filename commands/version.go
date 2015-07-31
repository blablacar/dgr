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
		buildDate, _ := cnt.BuildDate()
		fmt.Print("Cnt\n\n");
		fmt.Printf("version    : %s\n", cnt.Version)
		fmt.Printf("build date : %s\n", buildDate)
		if (cnt.CommitHash != "") {
			fmt.Printf("CommitHash : %s\n", cnt.CommitHash)
		}
		os.Exit(0)
	},
}
