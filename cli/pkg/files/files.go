package files

import (
	"os"
	"path"
	"strings"
)

func FileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, err
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
