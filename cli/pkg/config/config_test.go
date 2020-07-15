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
		Storage:            ".replicate/storage/",
	}, conf)

	// Loads a basic config
	err = ioutil.WriteFile(path.Join(dir, "replicate.yaml"), []byte("storage: 'foo'"), 0644)
	require.NoError(t, err)
	conf, _, err = FindConfig(dir)
	require.NoError(t, err)
	require.Equal(t, &Config{
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install:            []string{},
		Storage:            "foo",
	}, conf)
}

func TestParse(t *testing.T) {
	// Disallows unknown fields
	_, err := Parse([]byte("unknown: 'field'"))
	require.Error(t, err)

	// Sets defaults in empty config
	conf, err := Parse([]byte(""))
	require.NoError(t, err)
	require.Equal(t, &Config{
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install:            []string{},
		Storage:            ".replicate/storage/",
	}, conf)
}
