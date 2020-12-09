package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/require"

	"github.com/replicate/replicate/go/pkg/config"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/testutil"
)

func TestDiffSameExperiment(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(repo, workingDir)

	au := aurora.NewAurora(false)
	out := new(bytes.Buffer)
	err = printDiff(out, au, proj, "1e", "3c")
	require.NoError(t, err)
	actual := out.String()

	expected := `
Experiment
ID:                       1eeeeee                        1eeeeee

Params
(no difference)

Python Packages
(no difference)

Checkpoint
ID:                       2cccccc                        3cccccc
Created:                  Mon, 02 Jan 2006 23:00:05 +08  Mon, 02 Jan 2006 23:01:05 +08

Metrics
metric-1:                 0.01                           0.02

`
	actual = testutil.TrimRightLines(actual)
	expected = expected[1:]
	require.Equal(t, expected, actual)
}

func TestDiffDifferentExperiment(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(repo, workingDir)

	au := aurora.NewAurora(false)
	out := new(bytes.Buffer)
	err = printDiff(out, au, proj, "1e", "4c")
	require.NoError(t, err)
	actual := out.String()

	expected := `
Experiment
ID:                       1eeeeee                        2eeeeee
Command:                  train.py --gamma=1.2 -x
Created:                  Mon, 02 Jan 2006 22:54:05 +08  Mon, 02 Jan 2006 23:03:05 +08
Host:                     10.1.1.1                       10.1.1.2

Params
param-1:                  100                            200
param-3:                  (not set)                      hi

Python Packages
foo:                      1.2.3                          (not set)

Checkpoint
ID:                       2cccccc                        4cccccc
Created:                  Mon, 02 Jan 2006 23:00:05 +08  Mon, 02 Jan 2006 23:02:05 +08
Step:                     20                             5

Metrics
metric-1:                 0.01                           (not set)
metric-2:                 2                              (not set)
metric-3:                 (not set)                      0.5

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
