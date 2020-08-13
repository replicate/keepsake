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
	"replicate.ai/cli/pkg/testutil"
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
Commit:                   2cccccc                   3cccccc
Experiment:               1eeeeee                   1eeeeee

Params
(no difference)

Metrics
label-1:                  0.01                      0.02

Labels
(no difference)

`
	actual = testutil.TrimRightLines(actual)
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
Commit:                   2cccccc                   4cccccc
Experiment:               1eeeeee                   2eeeeee

Params
param-1:                  100                       200
param-3:                  (not set)                 hi

Metrics
label-1:                  0.01                      (not set)
label-3:                  (not set)                 0.5

Labels
label-2:                  2                         (not set)

`
	actual = testutil.TrimRightLines(actual)
	expected = expected[1:]
	require.Equal(t, expected, actual)
}

func TestMapString(t *testing.T) {
	// string pointer helpers
	baz := "baz"
	bop := "bop"

	// same
	require.Equal(t, map[string][]*string{}, mapString(map[string]string{
		"same": "in both",
	}, map[string]string{
		"same": "in both",
	}))

	// just in left
	require.Equal(t, map[string][]*string{
		"left": {&baz, nil},
	}, mapString(map[string]string{
		"same": "in both",
		"left": "baz",
	}, map[string]string{
		"same": "in both",
	}))

	// just in right
	require.Equal(t, map[string][]*string{
		"right": {nil, &baz},
	}, mapString(map[string]string{
		"same": "in both",
	}, map[string]string{
		"same":  "in both",
		"right": "baz",
	}))

	// different
	require.Equal(t, map[string][]*string{
		"different": {&baz, &bop},
	}, mapString(map[string]string{
		"same":      "in both",
		"different": "baz",
	}, map[string]string{
		"same":      "in both",
		"different": "bop",
	}))
}
