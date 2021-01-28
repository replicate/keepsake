package repository

import (
	"io/ioutil"
	"os"
	"path"
	"sort"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/replicate/keepsake/go/pkg/errors"
	"github.com/replicate/keepsake/go/pkg/files"
)

func TestDiskRepositoryGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	repository, err := NewDiskRepository(dir)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(dir, "some-file"), []byte("hello"), 0644)
	require.NoError(t, err)

	_, err = repository.Get("does-not-exist")
	require.True(t, errors.IsDoesNotExist(err))

	content, err := repository.Get("some-file")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)
}

func TestDiskGetPathTar(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	repository, err := NewDiskRepository(dir)
	require.NoError(t, err)

	tmpDir, err := files.TempDir("test")
	require.NoError(t, err)
	err = repository.GetPathTar("does-not-exist.tar.gz", tmpDir)
	require.True(t, errors.IsDoesNotExist(err))
}

func TestDiskGetPathItemTar(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create some temporary files
	fileDir := path.Join(dir, "files")
	err = os.MkdirAll(fileDir, os.ModePerm)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(fileDir, "a.txt"), []byte("file a"), 0644)
	require.NoError(t, err)
	err = ioutil.WriteFile(path.Join(fileDir, "b.txt"), []byte("file b"), 0644)
	require.NoError(t, err)
	err = os.Mkdir(path.Join(fileDir, "c"), 0755)
	require.NoError(t, err)
	err = ioutil.WriteFile(path.Join(fileDir, "c/d.txt"), []byte("file d"), 0644)
	require.NoError(t, err)

	repository, err := NewDiskRepository(dir)
	require.NoError(t, err)

	// Archive the sub-directory as a tarball in the repository
	// This should result in a tarball with the following directory tree:
	//
	// temp
	// |--c
	// |  |-- d.txt
	// |
	// |-- a.txt
	// |-- b.txt
	err = repository.PutPathTar(fileDir, "temp.tar.gz", "")
	require.NoError(t, err)

	// Create a temporary directory
	tmpDir, err := files.TempDir("test")
	require.NoError(t, err)

	// Extract just one of the two files from the repo dir.
	err = repository.GetPathItemTar("temp.tar.gz", "a.txt", tmpDir)
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(tmpDir, "a.txt"))
	require.NoError(t, err)
	require.Equal(t, []byte("file a"), content)

	// Extract an entire directory
	err = repository.GetPathItemTar("temp.tar.gz", "c", tmpDir)
	require.NoError(t, err)

	content, err = ioutil.ReadFile(path.Join(tmpDir, "c/d.txt"))
	require.NoError(t, err)
	require.Equal(t, []byte("file d"), content)

	// Extract a file that does not exist
	err = repository.GetPathItemTar("temp.tar.gz", "does-not-exist.txt", tmpDir)
	require.True(t, errors.IsDoesNotExist(err))
}

func TestDiskRepositoryPut(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	repository, err := NewDiskRepository(dir)
	require.NoError(t, err)

	err = repository.Put("some-file", []byte("hello"))
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(dir, "some-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)

	err = repository.Put("subdirectory/another-file", []byte("hello again"))
	require.NoError(t, err)

	content, err = ioutil.ReadFile(path.Join(dir, "subdirectory/another-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello again"), content)
}

func TestDiskRepositoryList(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	repository, err := NewDiskRepository(dir)
	require.NoError(t, err)

	err = repository.Put("some-file", []byte("hello"))
	require.NoError(t, err)
	err = repository.Put("dir/another-file", []byte("hello"))
	require.NoError(t, err)

	paths, err := repository.List("")
	require.NoError(t, err)
	require.Equal(t, []string{"some-file"}, paths)

	paths, err = repository.List("dir")
	require.NoError(t, err)
	require.Equal(t, []string{"dir/another-file"}, paths)

	paths, err = repository.List("dir-that-does-not-exist")
	require.NoError(t, err)
	require.Equal(t, []string{}, paths)
}

func TestDiskRepositoryListTarFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Create some temporary files to put inside a tarball.
	fileDir := path.Join(dir, "files")
	err = os.MkdirAll(fileDir, os.ModePerm)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(fileDir, "a.txt"), []byte("file a"), 0644)
	require.NoError(t, err)
	err = ioutil.WriteFile(path.Join(fileDir, "b.txt"), []byte("file b"), 0644)
	require.NoError(t, err)
	err = os.Mkdir(path.Join(fileDir, "c"), 0755)
	require.NoError(t, err)
	err = ioutil.WriteFile(path.Join(fileDir, "c/d.txt"), []byte("file d"), 0644)
	require.NoError(t, err)

	repository, err := NewDiskRepository(dir)
	require.NoError(t, err)

	// Archive the sub-directory as a tarball in the repository
	// This should result in a tarball with the following directory tree:
	//
	// temp
	// |--c
	// |  |-- d.txt
	// |
	// |-- a.txt
	// |-- b.txt
	err = repository.PutPathTar(fileDir, "temp.tar.gz", "")
	require.NoError(t, err)

	paths, err := repository.ListTarFile("temp.tar.gz")
	sort.Strings(paths)

	require.NoError(t, err)
	require.Equal(t, []string{"a.txt", "b.txt", "c/d.txt"}, paths)
}

func TestPutPath(t *testing.T) {
	repositoryDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(repositoryDir)

	repository, err := NewDiskRepository(repositoryDir)
	require.NoError(t, err)

	workDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(workDir)
	require.NoError(t, ioutil.WriteFile(path.Join(workDir, "some-file"), []byte("hello"), 0644))
	require.NoError(t, os.Mkdir(path.Join(workDir, "subdirectory"), 0755))
	require.NoError(t, ioutil.WriteFile(path.Join(workDir, "subdirectory/another-file"), []byte("hello again"), 0644))

	err = repository.PutPath(workDir, "parent")
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(repositoryDir, "parent/some-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)

	content, err = ioutil.ReadFile(path.Join(repositoryDir, "parent/subdirectory/another-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello again"), content)
}

func TestDiskListRecursive(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Works with emty repository
	repository, err := NewDiskRepository(dir)
	require.NoError(t, err)
	results := make(chan ListResult)
	go repository.ListRecursive(results, "checkpoints")
	require.Empty(t, <-results)

	// Lists stuff!
	require.NoError(t, repository.Put("checkpoints/abc123.json", []byte("yep")))
	require.NoError(t, repository.Put("experiments/def456.json", []byte("nope")))
	results = make(chan ListResult)
	go repository.ListRecursive(results, "checkpoints")
	require.Equal(t, ListResult{
		Path: "checkpoints/abc123.json",
		MD5:  []byte{0x93, 0x48, 0xae, 0x78, 0x51, 0xcf, 0x3b, 0xa7, 0x98, 0xd9, 0x56, 0x4e, 0xf3, 0x8, 0xec, 0x25},
	}, <-results)
	require.Empty(t, <-results)
}

func TestDiskMatchFilenamesRecursive(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Works with emty repository
	repository, err := NewDiskRepository(dir)
	require.NoError(t, err)
	results := make(chan ListResult)
	go repository.MatchFilenamesRecursive(results, "checkpoints", "keepsake-metadata.json")
	v := <-results
	require.Empty(t, v)
}
