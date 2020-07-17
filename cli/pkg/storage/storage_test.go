package storage

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

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

	err = PutDirectory(storage, "parent", workDir)
	require.NoError(t, err)

	content, err := ioutil.ReadFile(path.Join(storageDir, "parent/some-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), content)

	content, err = ioutil.ReadFile(path.Join(storageDir, "parent/subdirectory/another-file"))
	require.NoError(t, err)
	require.Equal(t, []byte("hello again"), content)
}
