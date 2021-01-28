package repository

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/replicate/keepsake/go/pkg/errors"
)

func TestSync(t *testing.T) {
	dir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	sourceRepository, err := NewDiskRepository(dir)
	require.NoError(t, err)

	dir, err = ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)
	destRepository, err := NewDiskRepository(dir)
	require.NoError(t, err)

	// Create files to dest various cases
	err = sourceRepository.Put("src-path/in-source-but-not-dest", []byte("hello"))
	require.NoError(t, err)

	err = destRepository.Put("dest-path/in-dest-but-not-in-source", []byte("bye"))
	require.NoError(t, err)

	err = sourceRepository.Put("src-path/same-content", []byte("hello"))
	require.NoError(t, err)
	err = destRepository.Put("dest-path/same-content", []byte("hello"))
	require.NoError(t, err)
	info, _ := os.Stat(filepath.Join(destRepository.rootDir, "dest-path/same-content"))
	sameContentMTime := info.ModTime()

	err = sourceRepository.Put("src-path/different-content", []byte("what is up"))
	require.NoError(t, err)
	err = destRepository.Put("dest-path/different-content", []byte("hello"))
	require.NoError(t, err)

	// Sync
	err = Sync(sourceRepository, "src-path", destRepository, "dest-path")
	require.NoError(t, err)

	// Test it was put in correct state
	data, err := destRepository.Get("dest-path/in-source-but-not-dest")
	require.NoError(t, err)
	require.Equal(t, []byte("hello"), data)

	_, err = destRepository.Get("dest-path/in-dest-but-not-in-source")
	require.True(t, errors.IsDoesNotExist(err))

	data, err = destRepository.Get("dest-path/different-content")
	require.NoError(t, err)
	require.Equal(t, []byte("what is up"), data)

	// Check same content hasn't been touched
	info, _ = os.Stat(filepath.Join(destRepository.rootDir, "dest-path/same-content"))
	require.Equal(t, info.ModTime(), sameContentMTime)
}
