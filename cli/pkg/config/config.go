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

const MaxSearchDepth = 100

type Config struct {
	Storage string `json:"storage"`
}

func FindConfig(folder string) (conf *Config, err error) {
	configPath, err := findConfigPath(folder)
	if err != nil {
		return nil, err
	}
	return LoadConfig(configPath)
}

func LoadConfig(configPath string) (conf *Config, err error) {
	text, err := ioutil.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read config file %s", configPath)
	}
	conf, err = Parse(text)
	if err != nil {
		// FIXME (bfirsh): implement standard way of displaying config errors so this can be used in other places
		msg := fmt.Sprintf("%v\n\n", err)
		msg += fmt.Sprintf("To fix this, take a look at the replicate.yaml reference:\n")
		msg += fmt.Sprintf("%s/docs/replicate-yaml", global.WebURL)
		return nil, fmt.Errorf(msg)
	}
	return conf, nil
}

func Parse(text []byte) (conf *Config, err error) {
	conf = new(Config)

	j, err := yaml.YAMLToJSON(text)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(bytes.NewReader(j))
	decoder.DisallowUnknownFields()
	err = decoder.Decode(&conf)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse replicate.yaml, got error: %s", err)
	}

	return conf, nil
}

func findConfigPath(startFolder string) (configPath string, err error) {
	folder := startFolder
	for i := 0; i < MaxSearchDepth; i++ {
		configPath = filepath.Join(folder, global.ConfigFilename)
		exists, err := files.FileExists(configPath)
		if err != nil {
			return "", fmt.Errorf("Failed to scan directory %s, got error: %s", folder, err)
		}
		if exists {
			return configPath, nil
		}

		if folder == "/" {
			return "", fmt.Errorf("%s not found in %s (or in any parent directories", global.ConfigFilename, startFolder)
		}

		folder = filepath.Dir(folder)
	}
	return "", fmt.Errorf("%s not found, recursive reached max depth", global.ConfigFilename)
}
