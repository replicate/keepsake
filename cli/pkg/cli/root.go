package cli

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/analytics"
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
	if userSettings.Email == "" && reportAnalytics {
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
		Short: "Reproducible packaging for maching learning models.",
		// Long: ``, <-- TODO

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
		newFeedbackCommand(),
	)

	return &rootCmd, nil
}

func ExecuteWithArgs(cmd *cobra.Command, args ...string) error {
	// HACK: no simple way of passing through variables for tests
	os.Setenv("REPLICATE_NO_ANALYTICS", "1")
	cmd.SetArgs(args)
	return cmd.Execute()
}

func setPersistentFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolVar(&global.Color, "color", true, "Display color in output")
	cmd.PersistentFlags().StringVarP(&global.SourceDirectory, "source-directory", "D", ".", "Local source directory")
	cmd.PersistentFlags().BoolVarP(&global.Verbose, "verbose", "v", false, "Verbose output")

}
