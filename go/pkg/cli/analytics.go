package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replicate/replicate/go/pkg/settings"
)

func newAnalyticsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "analytics <on|off>",
		Short: "Enable or disable analytics",
		Run:   handleErrors(analyticsCommand),
		Args:  cobra.ExactArgs(1),
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
