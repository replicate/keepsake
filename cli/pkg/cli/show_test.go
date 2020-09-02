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
		Created: fixedTime.Add(-10 * time.Minute),
		Params: map[string]*param.Value{
			"param-1": param.Int(100),
			"param-2": param.String("hello"),
		},
		Command: "train.py --gamma=1.2 -x",
		Host:    "10.1.1.1",
		User:    "andreas",
		Config:  conf,
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

	var checkpoints = []*project.Checkpoint{{
		ID:           "1ccccccccc",
		Created:      fixedTime.Add(-5 * time.Minute),
		ExperimentID: experiments[0].ID,
		Path:         "data",
		Metrics: map[string]*param.Value{
			"metric-1": param.Float(0.1),
			"metric-2": param.Int(2),
		},
		PrimaryMetric: &project.PrimaryMetric{
			Name: "metric-1",
			Goal: project.GoalMinimize,
		},
		Step: 10,
	}, {
		ID:           "2ccccccccc",
		Created:      fixedTime.Add(-4 * time.Minute),
		ExperimentID: experiments[0].ID,
		Path:         "data",
		Metrics: map[string]*param.Value{
			"metric-1": param.Float(0.01),
			"metric-2": param.Int(2),
		},
		PrimaryMetric: &project.PrimaryMetric{
			Name: "metric-1",
			Goal: project.GoalMinimize,
		},
		Step: 20,
	}, {
		ID:           "3ccccccccc",
		Created:      fixedTime.Add(-3 * time.Minute),
		ExperimentID: experiments[0].ID,
		Path:         "data",
		Metrics: map[string]*param.Value{
			"metric-1": param.Float(0.02),
			"metric-2": param.Int(2),
		},
		PrimaryMetric: &project.PrimaryMetric{
			Name: "metric-1",
			Goal: project.GoalMinimize,
		},
		Step: 20,
	}, {
		ID:           "4ccccccccc",
		Created:      fixedTime.Add(-2 * time.Minute),
		ExperimentID: experiments[1].ID,
		Path:         "data",
		Metrics: map[string]*param.Value{
			"metric-3": param.Float(0.5),
		},
		Step: 5,
	}}
	for _, com := range checkpoints {
		require.NoError(t, com.Save(store, workingDir))
	}

	require.NoError(t, project.CreateHeartbeat(store, experiments[0].ID, time.Now().UTC()))
	require.NoError(t, project.CreateHeartbeat(store, experiments[1].ID, time.Now().UTC().Add(-1*time.Minute)))

	return store
}

func TestShowCheckpoint(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	store := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(store)
	result, err := proj.CheckpointOrExperimentFromPrefix("3cc")
	require.NoError(t, err)
	require.NotNil(t, result.Checkpoint)

	out := new(bytes.Buffer)
	au := aurora.NewAurora(false)
	err = showCheckpoint(au, out, proj, result.Checkpoint)
	require.NoError(t, err)
	actual := out.String()

	expected := `
Checkpoint: 3ccccccccc

Created:    Mon, 02 Jan 2006 23:01:05 +08
Path:       data
Step:       20

Experiment
ID:         1eeeeeeeee
Created:    Mon, 02 Jan 2006 22:54:05 +08
Status:     running
Host:       10.1.1.1
User:       andreas
Command:    train.py --gamma=1.2 -x

Params
param-1:    100
param-2:    hello

Metrics
metric-1:   0.02 (primary, minimize)
metric-2:   2

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

	conf := &config.Config{}
	store := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(store)
	result, err := proj.CheckpointOrExperimentFromPrefix("1eee")
	require.NoError(t, err)
	require.NotNil(t, result.Experiment)

	out := new(bytes.Buffer)
	au := aurora.NewAurora(false)
	err = showExperiment(au, out, proj, result.Experiment)
	require.NoError(t, err)
	actual := out.String()

	expected := `
Experiment: 1eeeeeeeee

Created:  Mon, 02 Jan 2006 22:54:05 +08
Status:   running
Host:     10.1.1.1
User:     andreas
Command:  train.py --gamma=1.2 -x

Params
param-1:  100
param-2:  hello

Checkpoints
ID       STEP  CREATED     METRIC-1     METRIC-2
1cccccc  10    2006-01-02  0.1          2
2cccccc  20    2006-01-02  0.01 (best)  2
3cccccc  20    2006-01-02  0.02         2

To see more details about a checkpoint, run:
  replicate show <checkpoint ID>
`
	// remove initial newline
	expected = expected[1:]
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}
