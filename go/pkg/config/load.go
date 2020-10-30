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

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/files"
	"github.com/replicate/replicate/go/pkg/global"
)

const maxSearchDepth = 100
const deprecatedRepositoryDir = ".replicate/storage"

type configNotFoundError struct {
	message string
}

func (e *configNotFoundError) Error() string {
	return e.message + `
You must either create a replicate.yaml configuration file,
or explicitly pass the arguments 'repository' and 'directory' to
replicate.Project().
For more information, see https://beta.replicate.ai/docs"""
`
}

// FindConfigInWorkingDir searches working directory and any parent directories
// for replicate.yaml and loads it.
//
// This function can also be used to discover the source dir -- it returns a
// (config, projectDir) tuple.
//
// If overrideDir is passed, it uses that directory instead.
func FindConfigInWorkingDir(overrideDir string) (conf *Config, projectDir string, err error) {
	if overrideDir != "" {
		conf, err := LoadConfig(path.Join(overrideDir, global.ConfigFilename))
		if err != nil {
			if os.IsNotExist(err) {
				return getDefaultConfig(overrideDir), overrideDir, nil
			}
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
// directories for replicate.yaml, then loads it
func FindConfig(dir string) (conf *Config, projectDir string, err error) {
	configPath, deprecatedRepositoryProjectRoot, err := FindConfigPath(dir)
	if err != nil {
		return nil, "", err
	}
	if deprecatedRepositoryProjectRoot != "" {
		// go up two directories from .replicate/storage
		console.Warn(`replicate.yaml is required now. put this file in the project directory %s to remove this warning:

repository: file://.replicate/storage`, deprecatedRepositoryProjectRoot)

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

// LoadConfig reads and validates replicate.yaml
func LoadConfig(configPath string) (conf *Config, err error) {
	text, err := ioutil.ReadFile(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &configNotFoundError{fmt.Sprintf("The config path does not exist: %s", configPath)}
		}
		return nil, fmt.Errorf("Failed to read config file '%s': %w", configPath, err)
	}
	conf, err = Parse(text, path.Dir(configPath))
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
		return nil, fmt.Errorf("Failed to parse replicate.yaml: %s", err)
	}

	if conf.Storage != "" {
		// TODO(andreas): check that 'repository' and 'storage' aren't both defined (needs refactoring of defaults)
		console.Warn("'storage' is deprecated in replicate.yaml, please use 'repository'")
		conf.Repository = conf.Storage
		conf.Storage = ""
	}

	// TODO(andreas): generalize this once we have more required fields
	if conf.Repository == "" {
		return nil, fmt.Errorf("Missing required field in replicate.yaml: repository")
	}

	return conf, nil
}

func FindConfigPath(startFolder string) (configPath string, deprecatedRepositoryProjectRoot string, err error) {
	folder := startFolder
	for i := 0; i < maxSearchDepth; i++ {
		configPath = filepath.Join(folder, global.ConfigFilename)
		exists, err := files.FileExists(configPath)
		if err != nil {
			return "", "", fmt.Errorf("Failed to scan directory %s: %s", folder, err)
		}
		if exists {
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
			return "", "", &configNotFoundError{message: fmt.Sprintf("%s not found in %s (or in any parent directories", global.ConfigFilename, startFolder)}
		}

		folder = filepath.Dir(folder)
	}
	return "", "", &configNotFoundError{message: fmt.Sprintf("%s not found, recursive reached max depth", global.ConfigFilename)}
}
