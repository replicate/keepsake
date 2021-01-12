package files

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

const tempFolder = "/tmp/replicate"

func FileExists(filePath string) (bool, error) {
	if _, err := os.Stat(filePath); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("Failed to determine if %s exists: %w", filePath, err)
	}
}

func IsDir(dirPath string) (bool, error) {
	file, err := os.Stat(dirPath)
	if err != nil {
		return false, err
	}
	return file.Mode().IsDir(), nil
}

func TempDir(prefix string) (string, error) {
	// FIXME(bfirsh): make this more unique (e.g. ai.replicate, like some OS X applications do)

	err := os.MkdirAll(tempFolder, 0755)
	if err != nil {
		return "", fmt.Errorf("Failed to create temporary directory %s: %w", tempFolder, err)
	}
	name, err := ioutil.TempDir(tempFolder, prefix+"-")
	if err != nil {
		return "", fmt.Errorf("Failed to create temporary directory at %s: %w", tempFolder, err)
	}
	return name, nil
}

func DirIsEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		return false, fmt.Errorf("Failed to open %s: %w", dirPath, err)
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	if err != nil {
		return false, fmt.Errorf("Failed to read directory at %s: %w", dirPath, err)
	}
	return false, nil
}

func CopyFile(src string, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("Failed to open %s while copying to %s: %w", src, dest, err)
	}
	defer in.Close()

	out, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Failed to create %s while copying %s: %w", dest, src, err)
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return fmt.Errorf("Failed to copy %s to %s: %w", src, dest, err)
	}
	return out.Close()
}
