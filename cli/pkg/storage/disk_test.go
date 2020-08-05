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
	require.IsType(t, &NotExistError{}, err)

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

func TestDiskMatchFilenamesRecursive(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Works with emty storage
	storage, err := NewDiskStorage(dir)
	require.NoError(t, err)
	results := make(chan ListResult)
	go storage.MatchFilenamesRecursive(results, "commits", "replicate-metadata.json")
	v := <-results
	// FIXME (bfirsh): an empty struct is a bit of a weird way to indicate that there is nothing in the
	// channel. Maybe it should be sending *ListResult and nil indicates empty?
	require.Empty(t, v)
}
