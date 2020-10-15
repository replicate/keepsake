package storage

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver/v3"
	gitignore "github.com/sabhiram/go-gitignore"

	"replicate.ai/cli/pkg/files"
)

var maxWorkers = 128

type Scheme string

const (
	SchemeDisk Scheme = "file"
	SchemeS3   Scheme = "s3"
	SchemeGCS  Scheme = "gs"
)

type ListResult struct {
	Path  string
	MD5   []byte
	Error error
}

// Storage represents a blob store
//
// TODO: this interface needs trimming. A lot of things exist on this interface for the shared library with
// Python, but perhaps we could detatch that API from this. For example, this API could provide a GetPath with
// reader, then the shared API could add extracting from tarball on top of that.
type Storage interface {
	// A human-readable representation of this storage location. For example: s3://my-bucket/root
	RootURL() string

	RootExists() (bool, error)

	// Get data at path
	Get(path string) ([]byte, error)

	// GetPath recursively copies storageDir to localDir
	GetPath(storagePath, localPath string) error

	// GetPathTar extracts tarball `tarPath` to `localPath`
	//
	// The first component of the tarball is stripped. E.g. Extracting a tarball with `abc123/weights` in it to `/code` would create `/code/weights`.
	GetPathTar(tarPath, localPath string) error

	// Put data at path
	Put(path string, data []byte) error

	// PutPath recursively puts the local `localPath` directory into path `storagePath` on storage
	PutPath(localPath, storagePath string) error

	// PutPathTar recursively puts the local `localPath` directory into a tar.gz file `tarPath` on storage.
	// If `includePath` is set, only that will be included
	//
	// For example, PutPathTar("/code", "/tmp/abc123.tar.gz", "data") on these files:
	// - /code/train.py
	// - /code/data/weights
	// will result in a tarball containing:
	// - `abc123/data/weights`
	PutPathTar(localPath, tarPath, basePath string) error

	// Delete deletes path. If path is a directory, it recursively deletes
	// all everything under path
	Delete(path string) error

	// List files in a path non-recursively
	//
	// Returns a list of paths, prefixed with the given path, that can be passed straight to Get().
	// Directories are not listed.
	// If path does not exist, an empty list will be returned.
	List(path string) ([]string, error)

	// List files in a path recursively
	ListRecursive(results chan<- ListResult, folder string)

	MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string)

	// PrepareRunEnv prepares a replicate run, returning a list of
	// environment variables
	PrepareRunEnv() ([]string, error)
}

// SplitURL splits a storage URL into <scheme>://<path>
func SplitURL(storageURL string) (scheme Scheme, bucket string, root string, err error) {
	u, err := url.Parse(storageURL)
	if err != nil {
		return "", "", "", err
	}
	switch u.Scheme {
	case "":
		return SchemeDisk, "", u.Path, nil
	case "file":
		return SchemeDisk, "", u.Host + u.Path, nil
	case "s3":
		return SchemeS3, u.Host, strings.TrimPrefix(u.Path, "/"), nil
	case "gs":
		return SchemeGCS, u.Host, strings.TrimPrefix(u.Path, "/"), nil
	}
	return "", "", "", fmt.Errorf("Unknown storage backend: %s", u.Scheme)
}

func ForURL(storageURL string) (Storage, error) {
	scheme, bucket, root, err := SplitURL(storageURL)
	if err != nil {
		return nil, err
	}
	switch scheme {
	case SchemeDisk:
		return NewDiskStorage(root)
	case SchemeS3:
		return NewS3Storage(bucket, root)
	case SchemeGCS:
		return NewGCSStorage(bucket, root)
	}

	return nil, fmt.Errorf("Unknown storage backend: %s", scheme)
}

// FIXME: should we keep on doing this?
var putPathAlwaysIgnore = []string{".replicate", ".git", "venv", ".mypy_cache"}

type fileToPut struct {
	Source string
	Dest   string
	Info   os.FileInfo
}

func getListOfFilesToPut(localPath string, storagePath string) ([]fileToPut, error) {
	// Perhaps this should be configurable, or done at a higher-level? It seems odd this is done at such a low level.
	var ignore *gitignore.GitIgnore
	var err error
	ignoreFilePath := filepath.Join(localPath, ".replicateignore")
	if isDir, _ := files.IsDir(localPath); isDir {
		if exists, _ := files.FileExists(ignoreFilePath); exists {
			ignore, err = gitignore.CompileIgnoreFile(ignoreFilePath)
			if err != nil {
				return nil, err
			}
		}
	}

	result := []fileToPut{}
	err = filepath.Walk(localPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			for _, dir := range putPathAlwaysIgnore {
				if info.Name() == dir {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Strip local path
		relativePath, err := filepath.Rel(localPath, currentPath)
		if err != nil {
			return err
		}

		if ignore != nil && ignore.MatchesPath(relativePath) {
			return nil
		}

		result = append(result, fileToPut{
			Source: currentPath,
			Dest:   path.Join(storagePath, relativePath),
			Info:   info,
		})
		return nil
	})
	return result, err
}

func putPathTar(localPath string, out io.Writer, tarFileName string, includePath string) error {
	// archiver doesn't make it easy to include/exclude files, or write to a writer, so we have
	// to implement all this ourselves
	// TODO: adapt archiver so we can use its Archive() method with writers

	z := archiver.NewTarGz()
	if err := z.Create(out); err != nil {
		return err
	}
	defer z.Close()

	// Prefix all paths with name of tarball so it isn't a rude tarball
	destPath := filepath.Join(strings.TrimSuffix(tarFileName, ".tar.gz"), includePath)

	files, err := getListOfFilesToPut(filepath.Join(localPath, includePath), destPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		fh, err := os.Open(file.Source)
		if err != nil {
			return err
		}

		// write it to the archive
		err = z.Write(archiver.File{
			FileInfo: archiver.FileInfo{
				FileInfo:   file.Info,
				CustomName: file.Dest,
			},
			ReadCloser: fh,
		})
		fh.Close()
		if err != nil {
			return err
		}
	}
	// Explicitly call Close() on success to capture error.
	return z.Close()
}

func extractTar(tarPath, localPath string) error {
	tar := archiver.NewTarGz()
	tar.StripComponents = 1
	tar.OverwriteExisting = true
	return tar.Unarchive(tarPath, localPath)
}

// NeedsCaching returns true if the storage is slow and needs caching
func NeedsCaching(storage Storage) bool {
	_, isDiskStorage := storage.(*DiskStorage)
	return !isDiskStorage
}
