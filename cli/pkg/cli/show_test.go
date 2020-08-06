package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/storage"
	"replicate.ai/cli/pkg/testutil"
)

func init() {
	timezone, _ = time.LoadLocation("Asia/Ulaanbaatar")
}

func createShowTestData(t *testing.T, workingDir string, conf *config.Config) storage.Storage {
	store, err := storage.NewDiskStorage(path.Join(workingDir, ".replicate/storage"))
	require.NoError(t, err)

	fixedTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	require.NoError(t, err)
	var experiments = []*project.Experiment{{
		ID:      "1eeeeeeeee",
		Created: fixedTime,
		Params: map[string]*param.Value{
			"param-1": param.Int(100),
			"param-2": param.String("hello"),
		},
		Host:   "10.1.1.1",
		User:   "andreas",
		Config: conf,
	}, {
		ID:      "2eeeeeeeee",
		Created: fixedTime.Add(-1 * time.Minute),
		Params: map[string]*param.Value{
			"param-1": param.Int(200),
			"param-2": param.String("hello"),
			"param-3": param.String("hi"),
		},
		Host:   "10.1.1.2",
		User:   "andreas",
		Config: conf,
	}}
	for _, exp := range experiments {
		require.NoError(t, exp.Save(store))
	}

	var commits = []*project.Commit{{
		ID:           "1ccccccccc",
		Created:      fixedTime.Add(-1 * time.Minute),
		ExperimentID: experiments[0].ID,
		Labels: map[string]*param.Value{
			"label-1": param.Float(0.1),
			"label-2": param.Int(2),
		},
		Step: 10,
	}, {
		ID:           "2ccccccccc",
		Created:      fixedTime,
		ExperimentID: experiments[0].ID,
		Labels: map[string]*param.Value{
			"label-1": param.Float(0.01),
			"label-2": param.Int(2),
		},
		Step: 20,
	}, {
		ID:           "3ccccccccc",
		Created:      fixedTime,
		ExperimentID: experiments[0].ID,
		Labels: map[string]*param.Value{
			"label-1": param.Float(0.02),
			"label-2": param.Int(2),
		},
		Step: 20,
	}, {
		ID:           "4ccccccccc",
		Created:      fixedTime,
		ExperimentID: experiments[1].ID,
		Labels: map[string]*param.Value{
			"label-3": param.Float(0.5),
		},
		Step: 5,
	}}
	for _, com := range commits {
		require.NoError(t, com.Save(store, workingDir))
	}

	require.NoError(t, project.CreateHeartbeat(store, experiments[0].ID, time.Now().UTC()))
	require.NoError(t, project.CreateHeartbeat(store, experiments[1].ID, time.Now().UTC().Add(-1*time.Minute)))

	return store
}

func TestShowCommit(t *testing.T) {
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
	result, err := proj.CommitOrExperimentFromPrefix("3cc")
	require.NoError(t, err)
	require.NotNil(t, result.Commit)

	out := new(bytes.Buffer)
	au := aurora.NewAurora(false)
	err = showCommit(au, out, proj, result.Commit)
	require.NoError(t, err)
	actual := out.String()

	expected := `
Commit: 3ccccccccc

Created:    Mon, 02 Jan 2006 23:04:05 +08
Step:       20

Experiment
ID:         1eeeeeeeee
Created:    Mon, 02 Jan 2006 23:04:05 +08
Status:     running
Host:       10.1.1.1
User:       andreas

Params
param-1:    100
param-2:    hello

Metrics
label-1:    0.02 (primary, goal: minimize)
label-3:    (not set) (goal: minimize)

Labels
label-2:    2

`
	// remove initial newline
	expected = expected[1:]
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestShowExperiment(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "label-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}},
	}

	store := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(store)
	result, err := proj.CommitOrExperimentFromPrefix("1eee")
	require.NoError(t, err)
	require.NotNil(t, result.Experiment)

	out := new(bytes.Buffer)
	au := aurora.NewAurora(false)
	err = showExperiment(au, out, proj, result.Experiment)
	require.NoError(t, err)
	actual := out.String()

	expected := `
Experiment: 1eeeeeeeee

Created:  Mon, 02 Jan 2006 23:04:05 +08
Status:   running
Host:     10.1.1.1
User:     andreas

Params
param-1:  100
param-2:  hello

Commits
ID       STEP  CREATED     LABEL-1      LABEL-2
1cccccc  10    2006-01-02  0.1          2
3cccccc  20    2006-01-02  0.02         2
2cccccc  20    2006-01-02  0.01 (best)  2
`
	// remove initial newline
	expected = expected[1:]
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}
