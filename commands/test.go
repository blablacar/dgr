package commands

import (
	"github.com/blablacar/cnt/builder"
	"github.com/spf13/cobra"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "test image(s)",
	Long:  `test image(s)`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunTestType(".", buildArgs)
	},
}

func discoverAndRunTestType(path string, args builder.BuildArgs) {
	if cnt, err := builder.NewAci(path, args); err == nil {
		cnt.Test()
	} else if pod, err := builder.NewPod(path, args); err == nil {
		pod.Test()
	} else {
		panic("Cannot find cnt-manifest.yml")
	}
}

func init() {
	testCmd.Flags().BoolVarP(&buildArgs.NoTestFail, "no-test-fail", "T", false, "Fail if no tests found")
}
