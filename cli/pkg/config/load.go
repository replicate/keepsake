package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/ghodss/yaml"

	"replicate.ai/cli/pkg/files"
	"replicate.ai/cli/pkg/global"
)

const maxSearchDepth = 100

type configNotFoundError struct {
	message string
}

func (e *configNotFoundError) Error() string {
	return e.message
}

// FindConfig searches the current directory and any parent
// directories for replicate.yaml, then loads it
func FindConfig(folder string) (conf *Config, sourceDir string, err error) {
	configPath, err := FindConfigPath(folder)
	if err != nil {
		if _, ok := err.(*configNotFoundError); ok {
			return getDefaultConfig(), folder, nil
		}
		return nil, "", err
	}
	conf, err = LoadConfig(configPath)
	if err != nil {
		return nil, "", err
	}
	return conf, filepath.Dir(configPath), nil
}

// LoadConfig reads and validates replicate.yaml
func LoadConfig(configPath string) (conf *Config, err error) {
	exists, err := files.FileExists(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to determine if %s exists, got error: %s", configPath, err)
	}
	if !exists {
		return getDefaultConfig(), nil
	}

	text, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file %s", configPath)
	}
	conf, err = Parse(text)
	if err != nil {
		// FIXME (bfirsh): implement standard way of displaying config errors so this can be used in other places
		msg := fmt.Sprintf("%v\n\n", err)
		msg += "To fix this, take a look at the replicate.yaml reference:\n"
		msg += fmt.Sprintf("%s/docs/replicate-yaml", global.WebURL)
		return nil, fmt.Errorf(msg)
	}
	return conf, nil
}

// Parse replicate.yaml
func Parse(text []byte) (conf *Config, err error) {
	conf = getDefaultConfig()

	j, err := yaml.YAMLToJSON(text)
	if err != nil {
		return nil, err
	}
	// If it's an empty file, don't decode, otherwise we get this weird null object that isn't nil
	if string(j) == "null" {
		return conf, nil
	}

	decoder := json.NewDecoder(bytes.NewReader(j))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse replicate.yaml, got error: %s", err)
	}

	return conf, nil
}

func FindConfigPath(startFolder string) (configPath string, err error) {
	folder := startFolder
	for i := 0; i < maxSearchDepth; i++ {
		configPath = filepath.Join(folder, global.ConfigFilename)
		exists, err := files.FileExists(configPath)
		if err != nil {
			return "", fmt.Errorf("Failed to scan directory %s, got error: %s", folder, err)
		}
		if exists {
			return configPath, nil
		}

		if folder == "/" {
			// These error messages aren't used anywhere, but I've left them in in case this function is used elsewhere in the future
			return "", &configNotFoundError{message: fmt.Sprintf("%s not found in %s (or in any parent directories", global.ConfigFilename, startFolder)}
		}

		folder = filepath.Dir(folder)
	}
	return "", &configNotFoundError{message: fmt.Sprintf("%s not found, recursive reached max depth", global.ConfigFilename)}
}
