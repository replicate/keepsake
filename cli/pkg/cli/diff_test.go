package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/project"
)

func TestDiffSameExperiment(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "label-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}, {
			Name: "label-3",
			Goal: config.GoalMinimize,
		}},
	}

	store := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(store)

	au := aurora.NewAurora(false)
	out := new(bytes.Buffer)
	err = printDiff(out, au, proj, "1e", "3c")
	require.NoError(t, err)
	actual := out.String()

	expected := `
Checkpoint:               2ccccccccc                3ccccccccc
Experiment:               1eeeeeeeee                1eeeeeeeee

Params
(no difference)

Metrics
label-1:                  0.01                      0.02

Labels
(no difference)

`
	expected = expected[1:]
	require.Equal(t, expected, actual)
}

func TestDiffDifferentExperiment(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "label-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}, {
			Name: "label-3",
			Goal: config.GoalMinimize,
		}},
	}

	store := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(store)

	au := aurora.NewAurora(false)
	out := new(bytes.Buffer)
	err = printDiff(out, au, proj, "1e", "4c")
	require.NoError(t, err)
	actual := out.String()

	expected := `
Checkpoint:               2ccccccccc                4ccccccccc
Experiment:               1eeeeeeeee                2eeeeeeeee

Params
param-1:                  100                       200
param-3:                  (not set)                 hi

Metrics
label-1:                  0.01                      (not set)
label-3:                  (not set)                 0.5

Labels
label-2:                  2                         (not set)

`
	expected = expected[1:]
	require.Equal(t, expected, actual)
}
