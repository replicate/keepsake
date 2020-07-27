package storage

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type Scheme string

const (
	SchemeDisk Scheme = "file"
	SchemeS3   Scheme = "s3"
	SchemeGCS  Scheme = "gs"
)

type ListResult struct {
	Path  string
	Error error
}

type Storage interface {
	Get(path string) ([]byte, error)
	// GetMultiple files in whatever way is most efficient for that storage mechanism
	GetMultiple(path []string) (map[string][]byte, error)
	Put(path string, data []byte) error
	PutDirectory(path, source string) error
	MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string)
}

// SplitURL splits a storage URL into <scheme>://<path>
func SplitURL(storageURL string) (scheme Scheme, path string, err error) {
	urlRegex := regexp.MustCompile("^([^:]+)://(.+)$")
	matches := urlRegex.FindStringSubmatch(storageURL)
	if len(matches) == 0 {
		return SchemeDisk, storageURL, nil
	}

	path = matches[2]
	switch matches[1] {
	case "s3":
		return SchemeS3, path, nil
	case "gs":
		return SchemeGCS, path, nil
	case "file":
		return SchemeDisk, path, nil
	}
	return "", "", fmt.Errorf("Unknown storage backend: %s", matches[1])
}

func ForURL(storageURL string) (Storage, error) {
	scheme, path, err := SplitURL(storageURL)
	if err != nil {
		return nil, err
	}
	switch scheme {
	case SchemeS3:
		return NewS3Storage(path)
	case SchemeGCS:
		return NewGCSStorage(path)
	case SchemeDisk:
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

func parallelGet(storage Storage, paths []string) (map[string][]byte, error) {
	maxWorkers := int64(128)

	type Result struct {
		path string
		data []byte
	}
	results := make([]Result, len(paths))

	group, ctx := errgroup.WithContext(context.Background())
	group.Go(func() error {
		sem := semaphore.NewWeighted(maxWorkers)

		for i, p := range paths {
			i, p := i, p
			if err := sem.Acquire(ctx, 1); err != nil {
				return err
			}
			group.Go(func() error {
				defer sem.Release(1)
				data, err := storage.Get(p)
				if err != nil {
					return err
				}
				results[i] = Result{p, data}
				return nil
			})
		}
		return nil
	})
	if err := group.Wait(); err != nil {
		return nil, err
	}
	// Rejig the results into {path: data} map
	ret := map[string][]byte{}
	for _, result := range results {
		ret[result.path] = result.data
	}
	return ret, nil
}
