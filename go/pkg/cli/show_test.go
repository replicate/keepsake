package cli

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/require"

	"github.com/replicate/replicate/go/pkg/config"
	"github.com/replicate/replicate/go/pkg/param"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/repository"
	"github.com/replicate/replicate/go/pkg/testutil"
)

func init() {
	timezone, _ = time.LoadLocation("Asia/Ulaanbaatar")
}

func createShowTestData(t *testing.T, workingDir string, conf *config.Config) repository.Repository {
	repo, err := repository.NewDiskRepository(path.Join(workingDir, ".replicate"))
	require.NoError(t, err)

	fixedTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")

	require.NoError(t, err)
	var experiments = []*project.Experiment{{
		ID:      "1eeeeeeeee",
		Created: fixedTime.Add(-10 * time.Minute),
		Params: param.ValueMap{
			"param-1": param.Int(100),
			"param-2": param.String("hello"),
		},
		Command:        "train.py --gamma=1.2 -x",
		Host:           "10.1.1.1",
		User:           "andreas",
		Config:         conf,
		PythonPackages: map[string]string{"foo": "1.2.3"},
		Checkpoints: []*project.Checkpoint{
			{
				ID:      "1ccccccccc",
				Created: fixedTime.Add(-5 * time.Minute),
				Path:    "data",
				Metrics: param.ValueMap{
					"metric-1": param.Float(0.1),
					"metric-2": param.Int(2),
				},
				PrimaryMetric: &project.PrimaryMetric{
					Name: "metric-1",
					Goal: project.GoalMinimize,
				},
				Step: 10,
			}, {
				ID:      "2ccccccccc",
				Created: fixedTime.Add(-4 * time.Minute),
				Path:    "data",
				Metrics: param.ValueMap{
					"metric-1": param.Float(0.01),
					"metric-2": param.Int(2),
				},
				PrimaryMetric: &project.PrimaryMetric{
					Name: "metric-1",
					Goal: project.GoalMinimize,
				},
				Step: 20,
			}, {
				ID:      "3ccccccccc",
				Created: fixedTime.Add(-3 * time.Minute),
				Path:    "data",
				Metrics: param.ValueMap{
					"metric-1": param.Float(0.02),
					"metric-2": param.Int(2),
				},
				PrimaryMetric: &project.PrimaryMetric{
					Name: "metric-1",
					Goal: project.GoalMinimize,
				},
				Step: 20,
			},
		},
	}, {
		ID:      "2eeeeeeeee",
		Created: fixedTime.Add(-1 * time.Minute),
		Params: param.ValueMap{
			"param-1": param.Int(200),
			"param-2": param.String("hello"),
			"param-3": param.String("hi"),
		},
		Host:   "10.1.1.2",
		User:   "andreas",
		Config: conf,
		Checkpoints: []*project.Checkpoint{
			{
				ID:      "4ccccccccc",
				Created: fixedTime.Add(-2 * time.Minute),
				Path:    "data",
				Metrics: param.ValueMap{
					"metric-3": param.Float(0.5),
				},
				Step: 5,
			},
		},
	}}
	for _, exp := range experiments {
		require.NoError(t, exp.Save(repo))
	}

	require.NoError(t, project.CreateHeartbeat(repo, experiments[0].ID, time.Now().UTC()))
	require.NoError(t, project.CreateHeartbeat(repo, experiments[1].ID, time.Now().UTC().Add(-1*time.Minute)))

	return repo
}

func TestShowCheckpoint(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(repo)
	result, err := proj.CheckpointOrExperimentFromPrefix("3cc")
	require.NoError(t, err)
	require.NotNil(t, result.Checkpoint)

	out := new(bytes.Buffer)
	au := aurora.NewAurora(false)
	err = showCheckpoint(au, out, proj, result.Experiment, result.Checkpoint)
	require.NoError(t, err)
	actual := out.String()

	expected := `
Checkpoint: 3ccccccccc

Created:         Mon, 02 Jan 2006 23:01:05 +08
Path:            data
Step:            20

Experiment
ID:              1eeeeeeeee
Created:         Mon, 02 Jan 2006 22:54:05 +08
Status:          running
Host:            10.1.1.1
User:            andreas
Command:         train.py --gamma=1.2 -x

Params
param-1:         100
param-2:         hello

Python Packages
foo:             1.2.3

Metrics
metric-1:        0.02 (primary, minimize)
metric-2:        2

`
	// remove initial newline
	expected = expected[1:]
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)

	// json
	out = new(bytes.Buffer)
	err = show(showOpts{repositoryURL: "file://" + path.Join(workingDir, ".replicate"), json: true}, []string{"3ccc"}, out)
	require.NoError(t, err)
	var chkpt project.Checkpoint
	require.NoError(t, json.Unmarshal(out.Bytes(), &chkpt))
	require.Equal(t, "3ccccccccc", chkpt.ID)
}

func TestShowExperiment(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createShowTestData(t, workingDir, conf)
	proj := project.NewProject(repo)
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

Created:         Mon, 02 Jan 2006 22:54:05 +08
Status:          running
Host:            10.1.1.1
User:            andreas
Command:         train.py --gamma=1.2 -x

Params
param-1:         100
param-2:         hello

Python Packages
foo:             1.2.3

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

	// json
	out = new(bytes.Buffer)
	err = show(showOpts{repositoryURL: "file://" + path.Join(workingDir, ".replicate"), json: true}, []string{"1eee"}, out)
	require.NoError(t, err)
	var exp project.Experiment
	require.NoError(t, json.Unmarshal(out.Bytes(), &exp))
	require.Equal(t, "1eeeeeeeee", exp.ID)
	require.Equal(t, "1ccccccccc", exp.Checkpoints[0].ID)

}
