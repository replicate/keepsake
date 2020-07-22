package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
)

type ListResult struct {
	Path  string
	Error error
}

type Storage interface {
	Get(path string) ([]byte, error)
	Put(path string, data []byte) error
	PutDirectory(path, source string) error
	MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string)
}

func ForURL(storageURL string) (Storage, error) {
	urlRegex := regexp.MustCompile("^([^:]+)://(.+)$")
	matches := urlRegex.FindStringSubmatch(storageURL)
	if len(matches) == 0 {
		return NewDiskStorage(storageURL)
	}

	scheme := matches[1]
	path := matches[2]

	switch scheme {
	case "s3":
		return NewS3Storage(path)
	case "gs":
		return NewGCSStorage(path)
	case "file":
		return NewDiskStorage(path)
	}

	return nil, fmt.Errorf("Unknown storage backend: %s", scheme)
}

var putDirectorySkip = []string{".replicate", ".git", "venv", ".mypy_cache"}

type fileToPut struct {
	Source string
	Dest   string
}

func putDirectoryFiles(dest, source string) ([]fileToPut, error) {
	result := []fileToPut{}
	err := filepath.Walk(source, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			for _, dir := range putDirectorySkip {
				if info.Name() == dir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Strip local path
		relativePath, err := filepath.Rel(source, currentPath)
		if err != nil {
			return err
		}

		result = append(result, fileToPut{
			Source: currentPath,
			Dest:   path.Join(dest, relativePath),
		})
		return nil
	})
	return result, err
}
