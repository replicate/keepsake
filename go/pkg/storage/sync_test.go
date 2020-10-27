package storage

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSync(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	sourceStorage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	dir, err = ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	destStorage, err := NewDiskStorage(dir)
	require.NoError(t, err)

	// Create files to dest various cases
	err = sourceStorage.Put("src-path/in-source-but-not-dest", []byte("hello"))
	require.NoError(t, err)

	err = destStorage.Put("dest-path/in-dest-but-not-in-source", []byte("bye"))
	require.NoError(t, err)

	err = sourceStorage.Put("src-path/same-content", []byte("hello"))
	require.NoError(t, err)
	err = destStorage.Put("dest-path/same-content", []byte("hello"))
	require.NoError(t, err)
	info, _ := os.Stat(filepath.Join(destStorage.rootDir, "dest-path/same-content"))
	sameContentMTime := info.ModTime()

	err = sourceStorage.Put("src-path/different-content", []byte("what is up"))
	require.NoError(t, err)
	err = destStorage.Put("dest-path/different-content", []byte("hello"))
	require.NoError(t, err)

	// Sync
	err = Sync(sourceStorage, "src-path", destStorage, "dest-path")
	require.NoError(t, err)

	// Test it was put in correct state
	data, err := destStorage.Get("dest-path/in-source-but-not-dest")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)

	_, err = destStorage.Get("dest-path/in-dest-but-not-in-source")
	require.IsType(t, &DoesNotExistError{}, err)

	data, err = destStorage.Get("dest-path/different-content")
	require.NoError(t, err)
	require.Equal(t, []byte("what is up"), data)

	// Check same content hasn't been touched
	info, _ = os.Stat(filepath.Join(destStorage.rootDir, "dest-path/same-content"))
	require.Equal(t, info.ModTime(), sameContentMTime)
}
