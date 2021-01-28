package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replicate/keepsake/go/pkg/settings"
)

func newAnalyticsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analytics <on|off>",
		Short: "Enable or disable analytics",
		Long: `The Keepsake CLI sends anonymous analytics about commands you run.

The following data is sent on each command line invokation:
- A random token for the machine (e.g. "9f3027bb-0eb8-917d-e5bf-c6c1bdb1fd0a")
- The subcommand you ran, without any options or arguments
  (e.g. "keepsake run", not "keepsake run python secretproject.py")
- The Keepsake version (e.g. "1.0.0")
- Your CPU architecture (e.g. "amd64")
- Your operating system (e.g. "linux")

To learn more, please refer to https://keepsake.ai/docs/learn/analytics

These analytics really help us, and we'd really appreciate it if you left it on.
But, if you want to opt out, you can run this command:

keepsake analytics off
`,
		Run:  handleErrors(analyticsCommand),
		Args: cobra.ExactArgs(1),
	}
}

func analyticsCommand(cmd *cobra.Command, args []string) error {
	userSettings, err := settings.LoadUserSettings()
	if err != nil {
		return err
	}

	switch args[0] {
	case "on":
		userSettings.AnalyticsEnabled = true
	case "off":
		userSettings.AnalyticsEnabled = false
	default:
		return fmt.Errorf("You need to pass either 'on' or 'off' as an argument.")
	}

	if err := userSettings.Save(); err != nil {
		return err
	}

	return nil
}
