package settings

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
)

// TODO(bfirsh): perhaps this could be stored in local keychain

func secretsDir() (string, error) {
	settingsDir, err := UserSettingsDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(settingsDir, "secrets"), nil
}

// Get a secret, returning nil if it doesn't exist
//
// Note that the name is used as the filename of the secret file, so it must not contain special characters.
func GetSecret(name string) ([]byte, error) {
	dir, err := secretsDir()
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadFile(filepath.Join(dir, name))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return data, nil
}

// SetSecret writes a secret to the local secrets directory.
//
// Note that the name is used as the filename of the secret file, so it must not contain special characters.
func SetSecret(name string, data []byte) error {
	dir, err := secretsDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0700); err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(dir, name), data, 0600)
}
