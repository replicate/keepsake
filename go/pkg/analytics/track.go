package analytics

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/replicate/keepsake/go/pkg/console"
	"github.com/replicate/keepsake/go/pkg/global"
	"github.com/replicate/keepsake/go/pkg/settings"
)

// TrackCommand is the main entrypoint for analytics. This does all the checking
// for no tracking, the onboarding, and the actual tracking itself.
//
// This should be called after your command has finished, because it will sometimes flush data
// over the network.
//
// Any errors are assumed to be non-fatal and just sent to logs.
func TrackCommand(cmdName string) error {
	// Envvar bypasses everything
	if os.Getenv("REPLICATE_NO_ANALYTICS") != "" || os.Getenv("KEEPSAKE_NO_ANALYTICS") != "" {
		console.Debug("KEEPSAKE_NO_ANALYTICS set, not tracking")
		return nil
	}

	// Don't track analytics command so they can disable cleanly
	if cmdName == "analytics" {
		console.Debug("Not tracking analytics command")
		return nil
	}

	userSettings, err := settings.LoadUserSettings()
	if err != nil {
		return err
	}

	// Check this here because user might have already run `keepsake analytics off`, and we don't want to onboard if they have.
	// Default is true, so this isn't hit when file is first created
	if !userSettings.AnalyticsEnabled {
		console.Debug("Analytics disabled")
		return nil
	}

	// Check if AnalyticsEnabled is set, because user might have already run `keepsake analytics off`
	if !userSettings.FirstRun {
		userSettings.FirstRun = true
		userSettings.AnalyticsEnabled = true
		if err := userSettings.Save(); err != nil {
			return err
		}
		Onboarding()
		// Don't gather any metrics on first run
		return nil
	}

	settingsDir, err := settings.UserSettingsDir()
	if err != nil {
		return err
	}

	if err := settings.MaybeMoveDeprecatedUserSettingsDir(); err != nil {
		return err
	}

	client, err := NewClient(&Config{
		Dir:         filepath.Join(settingsDir, "analytics"),
		SegmentKey:  global.SegmentKey,
		AnonymousID: userSettings.AnalyticsID,
	})
	if err != nil {
		return err
	}
	if err := client.Track("Run Command", map[string]interface{}{
		// Update analytics.md in the docs if you add anything here
		"command":           cmdName,
		"keepsake_version":  global.Version,
		"replicate_version": global.Version,
		"operating_system":  runtime.GOOS,
		"cpu_architecture":  runtime.GOARCH,
	}); err != nil {
		return err
	}
	return client.ConditionalFlush(15, 5*time.Minute)
}

func Onboarding() {
	console.Info(`Keepsake gathers anonymous usage data. Lots of programs do this without telling you, so we want to make sure you know and can opt-out.
For more information, see ` + global.WebURL + `/docs/analytics

The analytics helps us make Keepsake better, so leaving it on is really appreciated, but if you want switch it off run:
  $ keepsake analytics off

We haven't gathered anything yet, so if you run that, nothing will have been sent to us.`)
	// Add a bit of space between the onboarding and rest of command
	fmt.Fprintln(os.Stderr)
}
