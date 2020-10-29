package repository

import (
	"crypto/md5"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/otiai10/copy"

	"github.com/replicate/replicate/go/pkg/files"
)

type DiskRepository struct {
	rootDir string
}

func NewDiskRepository(rootDir string) (*DiskRepository, error) {
	return &DiskRepository{
		rootDir: rootDir,
	}, nil
}

func (s *DiskRepository) RootURL() string {
	return "file://" + s.rootDir
}

// Get data at path
func (s *DiskRepository) Get(p string) ([]byte, error) {
	data, err := ioutil.ReadFile(path.Join(s.rootDir, p))
	if err != nil && os.IsNotExist(err) {
		return nil, &DoesNotExistError{msg: "Get: path does not exist: " + p}
	}
	return data, err
}

// GetPath recursively copies repoDir to localDir
func (s *DiskRepository) GetPath(repoDir string, localDir string) error {
	if err := copy.Copy(path.Join(s.rootDir, repoDir), localDir); err != nil {
		return fmt.Errorf("Failed to copy directory from %s to %s: %w", repoDir, localDir, err)
	}
	return nil
}

// GetPathTar extracts tarball `tarPath` to `localPath`
//
// See repository.go for full documentation.
func (s *DiskRepository) GetPathTar(tarPath, localPath string) error {
	fullTarPath := path.Join(s.rootDir, tarPath)
	exists, err := files.FileExists(fullTarPath)
	if err != nil {
		return err
	}
	if !exists {
		return &DoesNotExistError{msg: "GetPathTar: does not exist: " + fullTarPath}
	}
	return extractTar(fullTarPath, localPath)
}

// Put data at path
func (s *DiskRepository) Put(p string, data []byte) error {
	fullPath := path.Join(s.rootDir, p)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(fullPath, data, 0644)
}

// PutPath recursively puts the local `localPath` directory into path `repoPath` in the repository
func (s *DiskRepository) PutPath(localPath string, repoPath string) error {
	files, err := getListOfFilesToPut(localPath, repoPath)
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

// PutPathTar recursively puts the local `localPath` directory into a tar.gz file `tarPath` in the repository
// If `includePath` is set, only that will be included.
//
// See repository.go for full documentation.
func (s *DiskRepository) PutPathTar(localPath, tarPath, includePath string) error {
	if !strings.HasSuffix(tarPath, ".tar.gz") {
		return fmt.Errorf("PutPathTar: tarPath must end with .tar.gz")
	}

	fullPath := path.Join(s.rootDir, tarPath)
	err := os.MkdirAll(filepath.Dir(fullPath), 0755)
	if err != nil {
		return err
	}

	tarFile, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer tarFile.Close()

	if err := putPathTar(localPath, tarFile, filepath.Base(tarPath), includePath); err != nil {
		return err
	}

	// Explicitly call Close() on success to capture error
	return tarFile.Close()
}

// Delete deletes path. If path is a directory, it recursively deletes
// all everything under path
func (s *DiskRepository) Delete(pathToDelete string) error {
	if err := os.RemoveAll(path.Join(s.rootDir, pathToDelete)); err != nil {
		return fmt.Errorf("Failed to delete %s/%s: %w", s.rootDir, pathToDelete, err)
	}
	return nil
}

// List files in a path non-recursively
//
// Returns a list of paths, prefixed with the given path, that can be passed straight to Get().
// Directories are not listed.
// If path does not exist, an empty list will be returned.
func (s *DiskRepository) List(p string) ([]string, error) {
	files, err := ioutil.ReadDir(path.Join(s.rootDir, p))
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil
		}
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

func (s *DiskRepository) ListRecursive(results chan<- ListResult, folder string) {
	err := filepath.Walk(path.Join(s.rootDir, folder), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			relPath, err := filepath.Rel(s.rootDir, path)
			if err != nil {
				return err
			}

			md5sum, err := md5File(path)
			if err != nil {
				return err
			}
			results <- ListResult{Path: relPath, MD5: md5sum}
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

func (s *DiskRepository) MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string) {
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

func md5File(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return nil, err
	}
	return h.Sum(nil), nil
}
