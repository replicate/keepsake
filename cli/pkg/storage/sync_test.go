package storage

import (
	"io/ioutil"
	"os"
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

	err = sourceStorage.Put("path/file", []byte("hello"))
	require.NoError(t, err)

	err = destStorage.Put("path/new-nope", []byte("bye"))
	require.NoError(t, err)

	err = Sync(sourceStorage, "path", destStorage, "new-path")
	require.NoError(t, err)

	data, err := destStorage.Get("new-path/file")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)

	_, err = destStorage.Get("new-path/nope")
	require.IsType(t, &NotExistError{}, err)
}
