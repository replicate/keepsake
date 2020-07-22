package storage

import (
	"fmt"
	"io/ioutil"
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

// PutDirectory recursively puts the local `source` directory into path `dest` on storage
//
// Parallels Storage.put_directory in Python.
func PutDirectory(storage Storage, dest, source string) error {
	// TODO (bfirsh): support ignore, like in Python
	return filepath.Walk(source, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		data, err := ioutil.ReadFile(currentPath)
		if err != nil {
			return err
		}

		// Strip local path
		relativePath, err := filepath.Rel(source, currentPath)
		if err != nil {
			return err
		}

		return storage.Put(path.Join(dest, relativePath), data)
	})
}
