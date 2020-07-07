package storage

import (
	"fmt"
	"regexp"
)

type ListResult struct {
	Path  string
	Error error
}

type Storage interface {
	Get(path string) ([]byte, error)
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
	case "file":
		return NewDiskStorage(path)
	}

	return nil, fmt.Errorf("Unknown storage backend: %s", scheme)
}
