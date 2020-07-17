package files

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

const tempFolder = "/tmp/replicate"

func FileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("Failed to determine if %s exists, got error: %w", filePath, err)
	}
}

func IsDir(dirPath string) (bool, error) {
	file, err := os.Stat(dirPath)
	if err != nil {
		// FIXME: this might not necessarily mean directory doesn't exist, it might be a legit error
		return false, err
	}
	return file.Mode().IsDir(), nil
}

func ExpandUser(filePath string) (string, error) {
	if strings.HasPrefix(filePath, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		return path.Join(home, filePath[1:]), nil
	}
	return filePath, nil
}

func TempDir(prefix string) (string, error) {
	// TODO(andreas): make windows compatible

	err := os.MkdirAll(tempFolder, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create temporary directory %s, got error: %s", tempFolder, err)
	}
	name, err := ioutil.TempDir(tempFolder, prefix+"-")
	if err != nil {
		return "", fmt.Errorf("Failed to create temporary directory at %s, got error: %s", tempFolder, err)
	}
	return name, nil
}
