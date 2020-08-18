package cli

import (
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/analytics"
	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/global"
	"replicate.ai/cli/pkg/settings"
)

func NewRootCommand() (*cobra.Command, error) {
	reportAnalytics := os.Getenv("REPLICATE_NO_ANALYTICS") == ""

	userSettings, err := settings.LoadUserSettings()
	if err != nil {
		return nil, err
	}

	// On first launch, ask user for email
	if console.IsTerminal() && userSettings.Email == "" && reportAnalytics {
		if userSettings.Email, err = analytics.RunOnboarding(); err != nil {
			return nil, err
		}
		if err := userSettings.Save(); err != nil {
			return nil, err
		}
	}

	analyticsClient := analytics.NewClient(&analytics.Config{
		Email: userSettings.Email,
		// FIXME (bfirsh): use different key for development so we don't get junk data
		SegmentKey: "MKaYmSZ2hW6P8OegI9g0sufjZeUh28g7",
	})

	rootCmd := cobra.Command{
		Use:   "replicate",
		Short: "Version control for machine learning",
		// TODO: append getting started link to end of help text?
		Long: `Replicate: Version control for machine learning.

To learn how to get started, go to ` + global.WebURL + `/docs/tutorial`,

		Version:       global.Version,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if global.Verbose {
				console.SetLevel(console.DebugLevel)
			}
			console.SetColor(global.Color)

			if reportAnalytics {
				analyticsClient.Track("Run Command", map[string]string{
					"command": cmd.Name(),
					"args":    strings.Join(args, " "),
					"rawArgs": strings.Join(os.Args, " "),
				})
			}
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			analyticsClient.Close()
		},
	}
	setPersistentFlags(&rootCmd)

	rootCmd.AddCommand(
		newCheckoutCommand(),
		newDeleteCommand(),
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
	cmd.PersistentFlags().StringVarP(&global.SourceDirectory, "source-directory", "D", "", "Local source directory")
	cmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, "Verbose output")

}

// loadConfig loads config from global.SourceDirectory if it's
// defined, or searches recursively from cwd. If no replicate.yaml is
// found, it creates a default config.
func loadConfig() (conf *config.Config, sourceDir string, err error) {
	if global.SourceDirectory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return nil, "", err
		}
		conf, sourceDir, err := config.FindConfig(cwd)
		if err != nil {
			return nil, "", err
		}
		return conf, sourceDir, nil
	}
	conf, err = config.LoadConfig(path.Join(global.SourceDirectory, global.ConfigFilename))
	if err != nil {
		return nil, "", err
	}
	return conf, global.SourceDirectory, nil
}
