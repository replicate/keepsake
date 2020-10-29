package config

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFindConfig(t *testing.T) {
	// Loads default config if no config exists
	dir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	conf, _, err := FindConfig(dir)
	require.NoError(t, err)
	require.Equal(t, &Config{
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install:            []string{},
		Repository:         path.Join(dir, ".replicate/storage/"),
	}, conf)

	// Loads a basic config
	err = ioutil.WriteFile(path.Join(dir, "replicate.yaml"), []byte("repository: 'foo'"), 0644)
	require.NoError(t, err)
	conf, _, err = FindConfig(dir)
	require.NoError(t, err)
	require.Equal(t, &Config{
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install:            []string{},
		Repository:         "foo",
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
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install:            []string{},
		Repository:         "foo",
	}, conf)

	// Loads default config if override directory doesn't have replicate.yaml
	emptyDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(emptyDir)
	conf, _, err = FindConfigInWorkingDir(emptyDir)
	require.NoError(t, err)
	require.Equal(t, &Config{
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install:            []string{},
		Repository:         path.Join(emptyDir, ".replicate/storage/"),
	}, conf)
}

func TestParse(t *testing.T) {
	// Disallows unknown fields
	_, err := Parse([]byte("unknown: 'field'"), "")
	require.Error(t, err)

	// Sets defaults in empty config
	conf, err := Parse([]byte(""), "/foo")
	require.NoError(t, err)
	require.Equal(t, &Config{
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install:            []string{},
		Repository:         "/foo/.replicate/storage",
	}, conf)
}
