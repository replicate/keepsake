package config

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindConfig(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Loads a basic config
	err = ioutil.WriteFile(path.Join(dir, "replicate.yaml"), []byte("repository: 'foo'"), 0644)
	require.NoError(t, err)
	conf, _, err := FindConfig(dir)
	require.NoError(t, err)
	require.Equal(t, &Config{
		Repository: "foo",
	}, conf)
}

func TestFindConfigInWorkingDir(t *testing.T) {
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	// Uses override directory if that is passed
	err = ioutil.WriteFile(path.Join(dir, "replicate.yaml"), []byte("repository: 'foo'"), 0644)
	require.NoError(t, err)
	conf, _, err := FindConfigInWorkingDir(dir)
	require.NoError(t, err)
	require.Equal(t, &Config{
		Repository: "foo",
	}, conf)

	// Throw error if override directory doesn't have replicate.yaml
	emptyDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(emptyDir)
	_, _, err = FindConfigInWorkingDir(emptyDir)
	require.Error(t, err)
}

func TestParse(t *testing.T) {
	// Disallows unknown fields
	_, err := Parse([]byte("unknown: 'field'"), "")
	require.Error(t, err)

	// Load empty config
	conf, err := Parse([]byte(""), "/foo")
	require.NoError(t, err)
	require.Equal(t, &Config{}, conf)

	// Sets defaults in empty config
	conf, err = Parse([]byte("repository: s3://foobar"), "/foo")
	require.NoError(t, err)
	require.Equal(t, &Config{
		Repository: "s3://foobar",
	}, conf)

}

func TestStorageBackwardsCompatible(t *testing.T) {
	conf, err := Parse([]byte("storage: 's3://foobar'"), "")
	require.NoError(t, err)
	require.Equal(t, &Config{
		Repository: "s3://foobar",
	}, conf)
}

func TestDeprecatedRepositoryBackwardsCompatible(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	err = os.MkdirAll(filepath.Join(tmpDir, ".replicate/storage"), 0755)
	require.NoError(t, err)

	conf, projectDir, err := FindConfig(tmpDir)
	require.NoError(t, err)
	require.Equal(t, &Config{
		Repository: "file://.replicate/storage",
	}, conf)
	require.Equal(t, tmpDir, projectDir)
}
