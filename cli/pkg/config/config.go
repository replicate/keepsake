package config

import (
	"fmt"
	"io/ioutil"
	"path"
	"strings"

	"replicate.ai/cli/pkg/files"
)

type MetricGoal string

const (
	GoalMaximize MetricGoal = "maximize"
	GoalMinimize MetricGoal = "minimize"
)

type Metric struct {
	Goal    MetricGoal `json:"goal"`
	Primary bool       `json:"primary"`
}

// Config is replicate.yaml
type Config struct {
	Storage            string            `json:"storage"`
	Python             string            `json:"python"`
	CUDA               string            `json:"cuda"`
	PythonRequirements string            `json:"python_requirements"`
	Install            []string          `json:"install"`
	InstallScript      string            `json:"install_script"`
	Metrics            map[string]Metric `json:"metrics"`
}

// ReadPythonRequirements returns trimmed lines of text from
// conf.PythonRequirements, ignoring empty lines and comments
func (conf *Config) ReadPythonRequirements(sourceDir string) (lines []string, err error) {
	requirementsPath := path.Join(sourceDir, conf.PythonRequirements)
	exists, err := files.FileExists(requirementsPath)
	if err != nil {
		return nil, err
	}
	if !exists {
		return []string{}, nil
	}

	contents, err := ioutil.ReadFile(requirementsPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read %s, got error: %s", requirementsPath, err)
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

func (conf *Config) PrimaryMetric() (string, *Metric) {
	for name, metric := range conf.Metrics {
		if metric.Primary {
			return name, &metric
		}
	}
	return "", nil
}

func (conf *Config) HasPrimaryMetric() bool {
	_, m := conf.PrimaryMetric()
	return m != nil
}

func getDefaultConfig(workingDir string) *Config {
	// should match defaults in config.py
	return &Config{
		Storage:            path.Join(workingDir, ".replicate/storage/"),
		Python:             "3.7",
		PythonRequirements: "requirements.txt",
		Install:            []string{},
	}
}
