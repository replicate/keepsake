package cli

import (
	"bytes"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"text/tabwriter"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/storage"
)

func createTestData(t *testing.T, workingDir string, conf *config.Config) storage.Storage {
	store, err := storage.NewDiskStorage(path.Join(workingDir, ".replicate/storage"))
	require.NoError(t, err)

	require.NoError(t, err)
	var experiments = []*project.Experiment{{
		ID:      "1eeeeeeeee",
		Created: time.Now().UTC(),
		Params: map[string]*param.Value{
			"param-1": param.Int(100),
			"param-2": param.String("hello"),
		},
		Host:   "10.1.1.1",
		User:   "andreas",
		Config: conf,
	}, {
		ID:      "2eeeeeeeee",
		Created: time.Now().UTC().Add(-1 * time.Minute),
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
		Created:      time.Now().UTC().Add(-1 * time.Minute),
		ExperimentID: experiments[0].ID,
		Labels: map[string]*param.Value{
			"label-1": param.Float(0.1),
			"label-2": param.Int(2),
		},
		Step: 10,
	}, {
		ID:           "2ccccccccc",
		Created:      time.Now().UTC(),
		ExperimentID: experiments[0].ID,
		Labels: map[string]*param.Value{
			"label-1": param.Float(0.01),
			"label-2": param.Int(2),
		},
		Step: 20,
	}, {
		ID:           "3ccccccccc",
		Created:      time.Now().UTC(),
		ExperimentID: experiments[0].ID,
		Labels: map[string]*param.Value{
			"label-1": param.Float(0.02),
			"label-2": param.Int(2),
		},
		Step: 20,
	}, {
		ID:           "4ccccccccc",
		Created:      time.Now().UTC(),
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

	store := createTestData(t, workingDir, conf)
	proj := project.NewProject(store)
	result, err := proj.CommitOrExperimentFromPrefix("3cc")
	require.NoError(t, err)
	require.NotNil(t, result.Commit)

	out := new(bytes.Buffer)
	w := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)

	au := aurora.NewAurora(false)
	err = showCommit(au, w, proj, result.Commit)
	require.NoError(t, err)
	actual := out.String()

	expected := `
Commit:  3ccccccccc

Experiment:  1eeeeeeeee
Params
param-1:  100
param-2:  hello

Metrics
label-1:  0.02 (primary, goal: minimize)
label-3:  (none) (goal: minimize)
Labels
label-1:  0.02
label-2:  2

`
	// remove initial newline
	expected = expected[1:]
	require.Equal(t, expected, actual)
}
