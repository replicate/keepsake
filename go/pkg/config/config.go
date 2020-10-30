package config

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"github.com/replicate/replicate/go/pkg/files"
)

// Config is replicate.yaml
type Config struct {
	Repository         string   `json:"repository"`
	Python             string   `json:"python"`
	CUDA               string   `json:"cuda"`
	PythonRequirements string   `json:"python_requirements"`
	Install            []string `json:"install"`
	InstallScript      string   `json:"install_script"`

	Storage string `json:"storage"` // deprecated
}

// ReadPythonRequirements returns trimmed lines of text from
// conf.PythonRequirements, ignoring empty lines and comments
func (conf *Config) ReadPythonRequirements(projectDir string) (lines []string, err error) {
	requirementsPath := path.Join(projectDir, conf.PythonRequirements)
	exists, err := files.FileExists(requirementsPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}

	contents, err := ioutil.ReadFile(requirementsPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s: %s", requirementsPath, err)
	}

	lines = []string{}
	for _, line := range strings.Split(string(contents), "\n") {
		line = strings.TrimSpace(line)
		isComment := strings.HasPrefix(line, "#")
		if len(line) > 0 && !isComment {
			lines = append(lines, line)
		}
	}
	return lines, nil
}

func getDefaultConfig(workingDir string) *Config {
	// should match defaults in config.py
	return &Config{}
}
