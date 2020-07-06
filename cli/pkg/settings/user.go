package settings

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/files"
)

// UserSettings represents global user settings that span multiple projects
type UserSettings struct {
	Email string `json:"email"`
}

// LoadUserSettings loads the global user settings from disk, returning blank struct
// if no file exists
func LoadUserSettings() (*UserSettings, error) {
	settings := UserSettings{}

	exists, err := files.FileExists(userSettingsPath())
	if err != nil {
		return nil, err
	}
	if !exists {
		return &settings, nil
	}
	text, err := ioutil.ReadFile(userSettingsPath())
	if err != nil {
		console.Warn("Failed to read %s, got error: %s", userSettingsPath(), err)
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
	bytes, err := json.Marshal(s)
	if err != nil {
		return err
	}
	dir := filepath.Dir(userSettingsPath())
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}

	err = ioutil.WriteFile(userSettingsPath(), bytes, 0600)
	if err != nil {
		return err
	}
	return nil
}

func userSettingsPath() string {
	return os.ExpandEnv("$HOME/.config/replicate/settings.json")
}
