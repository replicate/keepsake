package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hashicorp/go-uuid"
	"github.com/mitchellh/go-homedir"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/files"
)

// UserSettings represents global user settings that span multiple projects
type UserSettings struct {
	FirstRun         bool   `json:"first_run"` // Set after first run
	AnalyticsEnabled bool   `json:"analytics_enabled"`
	AnalyticsID      string `json:"analytics_id"`
}

// LoadUserSettings loads the global user settings from disk, returning default struct
// if no file exists
func LoadUserSettings() (*UserSettings, error) {
	analyticsID, err := uuid.GenerateUUID()
	if err != nil {
		return nil, err
	}
	settings := UserSettings{
		AnalyticsID:      analyticsID,
		AnalyticsEnabled: true,
		FirstRun:         false,
	}
	settingsPath, err := userSettingsPath()
	if err != nil {
		return nil, err
	}

	exists, err := files.FileExists(settingsPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return &settings, nil
	}
	text, err := ioutil.ReadFile(settingsPath)
	if err != nil {
		console.Warn("Failed to read %s: %s", settingsPath, err)
		return &settings, nil
	}

	err = json.Unmarshal(text, &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// Save saves global user settings to disk
func (s *UserSettings) Save() error {
	settingsPath, err := userSettingsPath()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(s, "", " ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(settingsPath)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	err = ioutil.WriteFile(settingsPath, bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func UserSettingsDir() (string, error) {
	return homedir.Expand("~/.config/replicate")
}

func userSettingsPath() (string, error) {
	return homedir.Expand("~/.config/replicate/settings.json")
}
