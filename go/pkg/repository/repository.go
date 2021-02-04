package repository

import (
	"archive/tar"
	"fmt"
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/mholt/archiver/v3"
	gitignore "github.com/sabhiram/go-gitignore"

	"github.com/replicate/keepsake/go/pkg/console"
	"github.com/replicate/keepsake/go/pkg/errors"
	"github.com/replicate/keepsake/go/pkg/files"
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

// Repository represents a blob store
//
// TODO: this interface needs trimming. A lot of things exist on this interface for the shared library with
// Python, but perhaps we could detatch that API from this. For example, this API could provide a GetPath with
// reader, then the shared API could add extracting from tarball on top of that.
type Repository interface {
	// A human-readable representation of this repository location. For example: s3://my-bucket/root
	RootURL() string

	// Get data at path
	Get(path string) ([]byte, error)

	// GetPath recursively copies repoDir to localDir
	GetPath(repoPath, localPath string) error

	// GetPathTar extracts tarball `tarPath` to `localPath`
	//
	// The first component of the tarball is stripped. E.g. Extracting a tarball with `abc123/weights` in it to `/code` would create `/code/weights`.
	GetPathTar(tarPath, localPath string) error

	// GetPathItemTar extracts `itemPath` from tarball `tarPath` to `localPath`
	//
	// itemPath can be a single file or a directory.
	GetPathItemTar(tarPath, itemPath, localPath string) error

	// Put data at path
	Put(path string, data []byte) error

	// PutPath recursively puts the local `localPath` directory into path `repoPath` in the repository
	PutPath(localPath, repoPath string) error

	// PutPathTar recursively puts the local `localPath` directory into a tar.gz file `tarPath` in the repository.
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

	// List files in a tar-file
	//
	// Returns a list of paths, present inside the give tarfile, that can be passed straight to GetPathItemTar()
	// Directories are not listed.
	ListTarFile(path string) ([]string, error)

	// List files in a path recursively
	ListRecursive(results chan<- ListResult, folder string)

	MatchFilenamesRecursive(results chan<- ListResult, folder string, filename string)
}

// SplitURL splits a repository URL into <scheme>://<path>
func SplitURL(repositoryURL string) (scheme Scheme, bucket string, root string, err error) {
	u, err := url.Parse(repositoryURL)
	if err != nil {
		return "", "", "", err
	}
	switch u.Scheme {
	case "":
		return "", "", "", unknownRepositoryScheme("")
	case "file":
		return SchemeDisk, "", u.Host + u.Path, nil
	case "s3":
		return SchemeS3, u.Host, strings.TrimPrefix(u.Path, "/"), nil
	case "gs":
		return SchemeGCS, u.Host, strings.TrimPrefix(u.Path, "/"), nil
	}
	return "", "", "", unknownRepositoryScheme(u.Scheme)
}

func ForURL(repositoryURL string, projectDir string) (Repository, error) {
	scheme, bucket, root, err := SplitURL(repositoryURL)
	if err != nil {
		return nil, err
	}
	switch scheme {
	case SchemeDisk:
		if !filepath.IsAbs(root) {
			root = path.Join(projectDir, root)
		}
		return NewDiskRepository(root)
	case SchemeS3:
		return NewS3Repository(bucket, root)
	case SchemeGCS:
		return NewGCSRepository(bucket, root)
	}

	return nil, unknownRepositoryScheme(string(scheme))
}

// FIXME: should we keep on doing this?
var putPathAlwaysIgnore = []string{".keepsake", ".replicate", ".git", ".mypy_cache"}

type fileToPut struct {
	Source string
	Dest   string
	Info   os.FileInfo
}

func getListOfFilesToPut(localPath string, repoPath string) ([]fileToPut, error) {
	// Perhaps this should be configurable, or done at a higher-level? It seems odd this is done at such a low level.
	var ignore *gitignore.GitIgnore
	var err error
	ignoreFilePath := filepath.Join(localPath, ".keepsakeignore")
	deprecatedIgnoreFilePath := filepath.Join(localPath, ".replicateignore")
	if isDir, _ := files.IsDir(localPath); isDir {
		if exists, _ := files.FileExists(ignoreFilePath); exists {
			ignore, err = gitignore.CompileIgnoreFile(ignoreFilePath)
			if err != nil {
				return nil, err
			}
		} else if exists, _ := files.FileExists(deprecatedIgnoreFilePath); exists {
			ignore, err = gitignore.CompileIgnoreFile(deprecatedIgnoreFilePath)
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
			// TODO(andreas): should we always exclude virtualenvs?
			isVenv, err := isVirtualenvDir(currentPath)
			if err != nil {
				return err
			}
			if isVenv {
				return filepath.SkipDir
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
			Dest:   path.Join(repoPath, relativePath),
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
		return errors.WriteError(err.Error())
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
			return errors.WriteError(err.Error())
		}
	}
	// Explicitly call Close() on success to capture error.
	if err := z.Close(); err != nil {
		return errors.WriteError(err.Error())
	}
	return nil
}

func extractTar(tarPath, localPath string) error {
	tar := archiver.NewTarGz()
	tar.StripComponents = 1
	tar.OverwriteExisting = true
	return tar.Unarchive(tarPath, localPath)
}

func getListOfFilesInTar(tarPath string) ([]string, error) {
	result := []string{}

	t := archiver.NewTarGz()
	err := t.Walk(tarPath, func(f archiver.File) error {
		th, ok := f.Header.(*tar.Header)
		if !ok {
			return fmt.Errorf("expected header to be *tar.Header but was %T", f.Header)
		}

		result = append(result, th.Name)
		return nil
	})

	return result, err
}

func extractTarItem(tarPath, itemPath, localPath string) error {
	tarBaseName := filepath.Base(strings.TrimSuffix(tarPath, ".tar.gz"))
	fullItemPath := path.Join(tarBaseName, itemPath)

	filesInTar, err := getListOfFilesInTar(tarPath)
	if err != nil {
		return err
	}

	// Check if itemPath is inside the tar
	itemPathExists := false
	for _, fileInTar := range filesInTar {
		if strings.HasPrefix(fileInTar, fullItemPath) {
			itemPathExists = true
			break
		}
	}

	if !itemPathExists {
		return errors.DoesNotExist("Path does not exist inside the tarfile: " + itemPath)
	}

	tmpDir, err := files.TempDir("temp-extract-dir")
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	tar := archiver.NewTarGz()
	tar.StripComponents = 1
	tar.OverwriteExisting = true
	err = tar.Extract(tarPath, fullItemPath, tmpDir)
	if err != nil {
		return err
	}

	walkPath := path.Join(tmpDir, tarBaseName)
	err = filepath.Walk(walkPath, func(currentPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relativePath, err := filepath.Rel(walkPath, currentPath)
		if err != nil {
			return err
		}
		newPath := path.Join(localPath, relativePath)

		dir := filepath.Dir(newPath)

		exists, err := files.FileExists(dir)
		if err != nil {
			return err
		}

		if !exists {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("Failed to create directory %q: %w", dir, err)
			}
		}
		err = os.Rename(currentPath, newPath)
		if err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

// NeedsCaching returns true if the repository URL is slow and needs caching
func NeedsCaching(repositoryURL string) (bool, error) {
	scheme, _, _, err := SplitURL(repositoryURL)
	return scheme != "file", err
}

func unknownRepositoryScheme(scheme string) error {
	var message string
	if scheme == "" {
		message = "Missing repository scheme"
	} else {
		message = "Unknown repository scheme: " + scheme
	}
	return fmt.Errorf(message + `.

Make sure your repository URL starts with either 'file://', 's3://', or 'gs://'.
See the documentation for more details: https://keepsake.ai/docs/reference/yaml`)
}

func isVirtualenvDir(path string) (bool, error) {
	// TODO(andreas): this is maybe not super robust
	return files.FileExists(filepath.Join(path, "pyvenv.cfg"))
}

func CopyToTempDir(localPath string, includePath string) (tempDir string, err error) {
	// normalize path
	includePath = filepath.Join(includePath)

	console.Debug("Copying files to temporary directory")
	start := time.Now()

	tempDir, err = files.TempDir("copy-to-temp-dir")
	if err != nil {
		return "", err
	}

	// we first scan the whole repository to get the list of eligable files,
	// then copy the ones that match the includePath.
	// TODO(andreas): only scan files in the includePath
	filesToCopy, err := getListOfFilesToPut(localPath, tempDir)
	count := 0
	for _, file := range filesToCopy {

		// only include files in includePath
		relPath, err := filepath.Rel(localPath, file.Source)
		if err != nil {
			return "", err
		}
		if !(includePath == "." || relPath == includePath || strings.HasPrefix(relPath, includePath+"/")) {
			continue
		}

		dir := path.Dir(file.Dest)
		dirExists, err := files.FileExists(dir)
		if err != nil {
			return "", err
		}
		if !dirExists {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return "", fmt.Errorf("Failed to create directory %s: %v", dir, err)
			}
		}
		if err := files.CopyFile(file.Source, file.Dest); err != nil {
			return "", fmt.Errorf("Failed to copy %s to %s: %v", file.Source, file.Dest, err)
		}
		count += 1
	}

	console.Debug("Copied %d files to temporary directory (took %.3f seconds)", count, time.Since(start).Seconds())

	return tempDir, err
}
