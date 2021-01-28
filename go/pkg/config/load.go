package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/ghodss/yaml"

	"github.com/replicate/keepsake/go/pkg/console"
	"github.com/replicate/keepsake/go/pkg/errors"
	"github.com/replicate/keepsake/go/pkg/files"
	"github.com/replicate/keepsake/go/pkg/global"
	"github.com/replicate/keepsake/go/pkg/slices"
)

const maxSearchDepth = 100
const deprecatedRepositoryDir = ".replicate/storage"

// FindConfigInWorkingDir searches working directory and any parent directories
// for keepsake.yaml (or keepsake.yml) and loads it.
//
// This function can also be used to discover the source dir -- it returns a
// (config, projectDir) tuple.
//
// If overrideDir is passed, it uses that directory instead.
func FindConfigInWorkingDir(overrideDir string) (conf *Config, projectDir string, err error) {
	if overrideDir != "" {
		configPath, err := findConfigPathInDirectory(overrideDir)
		if err != nil {
			if errors.IsConfigNotFound(err) {
				return getDefaultConfig(overrideDir), overrideDir, nil
			}
			return nil, "", err
		}

		conf, err := LoadConfig(configPath)
		if err != nil {
			return nil, "", err
		}
		return conf, overrideDir, nil
	}
	cwd, err := os.Getwd()
	if err != nil {
		return nil, "", err
	}
	return FindConfig(cwd)
}

// FindConfig searches the given directory and any parent
// directories for keepsake.yaml, then loads it
func FindConfig(dir string) (conf *Config, projectDir string, err error) {
	configPath, deprecatedRepositoryProjectRoot, err := FindConfigPath(dir)
	if err != nil {
		return nil, "", err
	}
	if deprecatedRepositoryProjectRoot != "" {
		// go up two directories from .keepsake/storage
		console.Warn(`keepsake.yaml is required now. put this file in the project directory %s to remove this warning:

repository: file://%s`, projectDir, deprecatedRepositoryProjectRoot)

		conf = &Config{
			Repository: "file://" + deprecatedRepositoryDir,
		}
		return conf, deprecatedRepositoryProjectRoot, nil
	}
	conf, err = LoadConfig(configPath)
	if err != nil {
		return nil, "", err
	}
	return conf, filepath.Dir(configPath), nil
}

// LoadConfig reads and validates keepsake.yaml
func LoadConfig(configPath string) (conf *Config, err error) {
	text, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.ConfigNotFound("The config path does not exist: " + configPath)
		}
		return nil, fmt.Errorf("Failed to read config file '%s': %w", configPath, err)
	}
	conf, err = Parse(text, path.Dir(configPath))
	if err != nil {
		// FIXME (bfirsh): implement standard way of displaying config errors so this can be used in other places
		msg := fmt.Sprintf("%v\n\n", err)
		msg += "To fix this, take a look at the keepsake.yaml reference:\n"
		msg += fmt.Sprintf("%s/docs/reference/yaml", global.WebURL)
		return nil, fmt.Errorf(msg)
	}
	return conf, nil
}

// Parse keepsake.yaml
func Parse(text []byte, dir string) (conf *Config, err error) {
	conf = getDefaultConfig(dir)

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
		return nil, fmt.Errorf("Failed to parse keepsake.yaml: %s", err)
	}

	if conf.Storage != "" {
		if conf.Repository != "" {
			return nil, fmt.Errorf("'repository' and 'storage' (deprecated) cannot both be defined, please only use 'repository'")
		}

		console.Warn("'storage' is deprecated in keepsake.yaml, please use 'repository'")
		conf.Repository = conf.Storage
		conf.Storage = ""
	}

	if conf.Repository == "" {
		return nil, fmt.Errorf("Missing required field in keepsake.yaml: repository")
	}

	return conf, nil
}

func FindConfigPath(startFolder string) (configPath string, deprecatedRepositoryProjectRoot string, err error) {
	folder := startFolder
	for i := 0; i < maxSearchDepth; i++ {
		configPath, err := findConfigPathInDirectory(folder)
		if err != nil && !errors.IsConfigNotFound(err) {
			return "", "", err
		}
		if err == nil {
			return configPath, "", nil
		}

		deprecatedRepo := filepath.Join(folder, deprecatedRepositoryDir)
		deprecatedRepositoryExists, err := files.FileExists(deprecatedRepo)
		if err != nil {
			return "", "", fmt.Errorf("Failed to scan directory %s: %s", folder, err)
		}
		if deprecatedRepositoryExists {
			return "", folder, nil
		}

		if folder == "/" {
			// These error messages aren't used anywhere, but I've left them in in case this function is used elsewhere in the future
			return "", "", errors.ConfigNotFound(fmt.Sprintf("%s not found in %s (or in any parent directories", global.ConfigFilenames[0], startFolder))
		}

		folder = filepath.Dir(folder)
	}
	return "", "", errors.ConfigNotFound(fmt.Sprintf("%s not found, recursive reached max depth", global.ConfigFilenames[0]))
}

func findConfigPathInDirectory(folder string) (configPath string, err error) {
	for _, configFilename := range global.ConfigFilenames {
		configPath = filepath.Join(folder, configFilename)
		exists, err := files.FileExists(configPath)
		if err != nil {
			return "", fmt.Errorf("Failed to scan directory %s: %s", folder, err)
		}
		if exists {
			if slices.ContainsString(global.DeprecatedConfigFilenames, configFilename) {
				console.Warn("%s is deprecated, please name your configuration file %s", configFilename, global.ConfigFilenames[0])
			}

			return configPath, nil
		}
	}
	return "", errors.ConfigNotFound(fmt.Sprintf("%s not found in %s", global.ConfigFilenames[0], folder))
}
