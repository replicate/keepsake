package storage

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/otiai10/copy"
)

type DiskStorage struct {
	rootDir string
}

func NewDiskStorage(rootDir string) (*DiskStorage, error) {
	return &DiskStorage{
		rootDir: rootDir,
	}, nil
}

// Get data at path
func (s *DiskStorage) Get(p string) ([]byte, error) {
	data, err := ioutil.ReadFile(path.Join(s.rootDir, p))
	if err != nil && os.IsNotExist(err) {
		return nil, &NotExistError{msg: "Get: path does not exist: " + p}
	}
	return data, err
}

// Put data at path
func (s *DiskStorage) Put(p string, data []byte) error {
	fullPath := path.Join(s.rootDir, p)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fullPath, data, 0644)
}

// PutDirectory recursively puts the local `localPath` directory into path `storagePath` on storage
//
// Parallels Storage.put_directory in Python.
func (s *DiskStorage) PutDirectory(localPath string, storagePath string) error {
	files, err := putDirectoryFiles(localPath, storagePath)
	if err != nil {
		return err
	}
	for _, file := range files {
		data, err := ioutil.ReadFile(file.Source)
		if err != nil {
			return err
		}
		err = s.Put(file.Dest, data)
		if err != nil {
			return err
		}
	}
	return nil
}

// List files in a path non-recursively
//
// Returns a list of paths, prefixed with the given path, that can be passed straight to Get().
// Directories are not listed.
// If path does not exist, an empty list will be returned.
func (s *DiskStorage) List(p string) ([]string, error) {
	files, err := ioutil.ReadDir(path.Join(s.rootDir, p))
	if err != nil {
		return nil, err
	}
	result := []string{}
	for _, f := range files {
		if !f.IsDir() {
			result = append(result, path.Join(p, f.Name()))
		}
	}
	return result, nil
}

func (s *DiskStorage) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
	err := filepath.Walk(path.Join(s.rootDir, folder), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if filepath.Base(path) == filename {
			relPath, err := filepath.Rel(s.rootDir, path)
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

// GetDirectory recursively copies storageDir to localDir
func (s *DiskStorage) GetDirectory(storageDir string, localDir string) error {
	if err := copy.Copy(path.Join(s.rootDir, storageDir), localDir); err != nil {
		return fmt.Errorf("Failed to copy directory from %s to %s, got error: %w", storageDir, localDir, err)
	}
	return nil
}
