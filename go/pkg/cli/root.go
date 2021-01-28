package cli

import (
	"github.com/spf13/cobra"

	"github.com/replicate/keepsake/go/pkg/analytics"
	"github.com/replicate/keepsake/go/pkg/console"
	"github.com/replicate/keepsake/go/pkg/global"
)

func NewRootCommand() (*cobra.Command, error) {
	rootCmd := cobra.Command{
		Use:   "keepsake",
		Short: "Version control for machine learning",
		// TODO: append getting started link to end of help text?
		Long: `Keepsake: Version control for machine learning.

To learn how to get started, go to ` + global.WebURL + `/docs/tutorial`,

		Version: global.Version,
		// This stops errors being printed because we print them in cmd/keepsake/main.go
		SilenceErrors: true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if global.Verbose {
				console.SetLevel(console.DebugLevel)
			}
			console.SetColor(global.Color)

			if err := analytics.TrackCommand(cmd.Name()); err != nil {
				console.Debug("analytics error: %s", err)
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
		},
	}
	setPersistentFlags(&rootCmd)

	rootCmd.AddCommand(
		newAnalyticsCommand(),
		newCheckoutCommand(),
		newRmCommand(),
		newDiffCommand(),
		newFeedbackCommand(),
		newGenerateDocsCommand(&rootCmd),
		newListCommand(),
		newPsCommand(),
		newShowCommand(),
	)

	return &rootCmd, nil
}

func setPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&global.Color, "color", true, "Display color in output")
	// FIXME (bfirsh): this noun needs standardizing. we use the term "working directory" in some places.
	cmd.PersistentFlags().StringVarP(&global.ProjectDirectory, "project-directory", "D", "", "Project directory. Default: nearest parent directory with keepsake.yaml")
	cmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, "Verbose output")

}
