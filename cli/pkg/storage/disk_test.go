package storage

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDiskStorageGet(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(dir, "some-file"), []byte("hello"), 0644)
	require.NoError(t, err)

	_, err = storage.Get("does-not-exist")
	require.IsType(t, &DoesNotExistError{}, err)

	content, err := storage.Get("some-file")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)
}

func TestDiskStoragePut(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	err = storage.Put("some-file", []byte("hello"))
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(dir, "some-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)

	err = storage.Put("subdirectory/another-file", []byte("hello again"))
	require.NoError(t, err)

	content, err = ioutil.ReadFile(path.Join(dir, "subdirectory/another-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello again"), content)
}

func TestDiskStorageList(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	err = storage.Put("some-file", []byte("hello"))
	require.NoError(t, err)
	err = storage.Put("dir/another-file", []byte("hello"))
	require.NoError(t, err)

	paths, err := storage.List("")
	require.NoError(t, err)
	require.Equal(t, []string{"some-file"}, paths)

	paths, err = storage.List("dir")
	require.NoError(t, err)
	require.Equal(t, []string{"dir/another-file"}, paths)

	paths, err = storage.List("dir-that-does-not-exist")
	require.NoError(t, err)
	require.Equal(t, []string{}, paths)
}

func TestPutDirectory(t *testing.T) {
	storageDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(storageDir)

	storage, err := NewDiskStorage(storageDir)
	require.NoError(t, err)

	workDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workDir)
	require.NoError(t, ioutil.WriteFile(path.Join(workDir, "some-file"), []byte("hello"), 0644))
	require.NoError(t, os.Mkdir(path.Join(workDir, "subdirectory"), 0755))
	require.NoError(t, ioutil.WriteFile(path.Join(workDir, "subdirectory/another-file"), []byte("hello again"), 0644))

	err = storage.PutDirectory(workDir, "parent")
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(storageDir, "parent/some-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)

	content, err = ioutil.ReadFile(path.Join(storageDir, "parent/subdirectory/another-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello again"), content)
}

func TestDiskListRecursive(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Works with emty storage
	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)
	results := make(chan ListResult)
	go storage.ListRecursive(results, "checkpoints")
	require.Empty(t, <-results)

	// Lists stuff!
	require.NoError(t, storage.Put("checkpoints/abc123.json", []byte("yep")))
	require.NoError(t, storage.Put("experiments/def456.json", []byte("nope")))
	results = make(chan ListResult)
	go storage.ListRecursive(results, "checkpoints")
	require.Equal(t, ListResult{Path: "checkpoints/abc123.json"}, <-results)
	require.Empty(t, <-results)
}

func TestDiskMatchFilenamesRecursive(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Works with emty storage
	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)
	results := make(chan ListResult)
	go storage.MatchFilenamesRecursive(results, "checkpoints", "replicate-metadata.json")
	v := <-results
	require.Empty(t, v)
}
