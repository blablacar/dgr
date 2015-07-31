package commands
import "github.com/spf13/cobra"


var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "push image(s)",
	Long:  `push images to repository`,
	Run: func(cmd *cobra.Command, args []string) {
		runCleanIfRequested(".", buildArgs)
		discoverAndRunPushType(".", buildArgs)
	},
}
