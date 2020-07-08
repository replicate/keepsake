package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/files"
)

type DiskStorage struct {
	folder string
}

func NewDiskStorage(folder string) (*DiskStorage, error) {
	return &DiskStorage{
		folder: folder,
	}, nil
}

func (s *DiskStorage) Get(p string) ([]byte, error) {
	return ioutil.ReadFile(path.Join(s.folder, p))
}

func (s *DiskStorage) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	err := filepath.Walk(path.Join(s.folder, folder), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == filename {
			relPath, err := filepath.Rel(s.folder, path)
			if err != nil {
				return err
			}
			results <- ListResult{Path: relPath}
		}
		return nil
	})
	if err != nil {
		// If directory does not exist, treat this as empty. This is consistent with how blob storage
		// would behave
		if os.IsNotExist(err) {
			close(results)
			return
		}

		results <- ListResult{Error: err}
	}
	close(results)
}

func (s *DiskStorage) ensureFolderExists() error {
	exists, err := files.FileExists(s.folder)
	if err != nil {
		return err
	}
	if exists {
		isDir, err := files.IsDir(s.folder)
		if err != nil {
			return err
		}
		if isDir {
			return nil
		}
		return fmt.Errorf("Storage path %s is not a directory", s.folder)
	}
	console.Debug("Creating disk storage folder: %s", s.folder)
	if err := os.MkdirAll(s.folder, 0755); err != nil {
		return fmt.Errorf("Failed to create folder: %s", s.folder)
	}

	return nil
}
