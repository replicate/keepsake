package storage

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

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
