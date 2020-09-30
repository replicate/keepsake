package cli

import (
	"os"
	"path"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/analytics"
	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/global"
)

func NewRootCommand() (*cobra.Command, error) {
	rootCmd := cobra.Command{
		Use:   "replicate",
		Short: "Version control for machine learning",
		// TODO: append getting started link to end of help text?
		Long: `Replicate: Version control for machine learning.

To learn how to get started, go to ` + global.WebURL + `/docs/tutorial`,

		Version: global.Version,
		// This stops errors being printed because we print them in cmd/replicate/main.go
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
		newRunCommand(),
		newShowCommand(),
	)

	return &rootCmd, nil
}

func ExecuteWithArgs(cmd *cobra.Command, args ...string) error {
	cmd.SetArgs(args)
	return cmd.Execute()
}

func setPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&global.Color, "color", true, "Display color in output")
	// FIXME (bfirsh): this noun needs standardizing. we use the term "working directory" in some places.
	cmd.PersistentFlags().StringVarP(&global.ProjectDirectory, "project-directory", "D", "", "Project directory. Default: working directory, or nearest parent with replicate.yaml")
	cmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, "Verbose output")

}

// loadConfig loads config from global.ProjectDirectory if it's
// defined, or searches recursively from cwd. If no replicate.yaml is
// found, it creates a default config.
func loadConfig() (conf *config.Config, projectDir string, err error) {
	if global.ProjectDirectory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", err
		}
		conf, projectDir, err := config.FindConfig(cwd)
		if err != nil {
			return nil, "", err
		}
		return conf, projectDir, nil
	}
	conf, err = config.LoadConfig(path.Join(global.ProjectDirectory, global.ConfigFilename))
	if err != nil {
		return nil, "", err
	}
	return conf, global.ProjectDirectory, nil
}
