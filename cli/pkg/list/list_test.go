package list

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/commit"
	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

func TestListJSON(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	storageDir := path.Join(workingDir, ".replicate/storage")

	storage, err := storage.NewDiskStorage(storageDir)
	require.NoError(t, err)
	defer os.RemoveAll(storageDir)

	// Experiment no longer running
	exp := experiment.NewExperiment(map[string]*param.Value{
		"learning_rate": param.Float(0.001),
	})
	require.NoError(t, exp.Save(storage))
	require.NoError(t, err)
	require.NoError(t, experiment.CreateHeartbeat(storage, exp.ID, time.Now().Add(-24*time.Hour)))
	com := commit.NewCommit(exp.ID, map[string]*param.Value{
		"accuracy": param.Float(0.987),
	})
	require.NoError(t, com.Save(storage, workingDir))

	// Experiment still running
	exp = experiment.NewExperiment(map[string]*param.Value{
		"learning_rate": param.Float(0.002),
	})
	require.NoError(t, exp.Save(storage))
	require.NoError(t, err)
	require.NoError(t, experiment.CreateHeartbeat(storage, exp.ID, time.Now()))
	com = commit.NewCommit(exp.ID, map[string]*param.Value{
		"accuracy": param.Float(0.987),
	})
	require.NoError(t, com.Save(storage, workingDir))

	// replicate list
	out := new(bytes.Buffer)
	err = Experiments(out, storage, FormatJSON)
	require.NoError(t, err)

	experiments := make([]GroupedExperiment, 0)
	dec := json.NewDecoder(out)
	require.NoError(t, dec.Decode(&experiments))
	require.Equal(t, 2, len(experiments))

	require.Equal(t, param.Float(0.001), experiments[0].Params["learning_rate"])
	require.Equal(t, 1, experiments[0].NumCommits)
	require.Equal(t, param.Float(0.987), experiments[0].LatestCommit.Metrics["accuracy"])
	require.Equal(t, false, experiments[0].Running)

	require.Equal(t, param.Float(0.002), experiments[1].Params["learning_rate"])
	require.Equal(t, 1, experiments[1].NumCommits)
	require.Equal(t, param.Float(0.987), experiments[1].LatestCommit.Metrics["accuracy"])
	require.Equal(t, true, experiments[1].Running)
}
