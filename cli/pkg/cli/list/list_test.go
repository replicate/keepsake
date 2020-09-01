package list

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/storage"
	"replicate.ai/cli/pkg/testutil"
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
		Host:    "10.1.1.1",
		User:    "andreas",
		Config:  conf,
		Command: "train.py --foo bar",
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
	}, {
		ID:      "3eeeeeeeee",
		Created: time.Now().UTC().Add(-2 * time.Minute),
		Params: map[string]*param.Value{
			"param-1": param.Int(200),
			"param-2": param.String("hello"),
			"param-3": param.String("hi"),
		},
		Host:   "10.1.1.2",
		User:   "ben",
		Config: conf,
	}}
	for _, exp := range experiments {
		require.NoError(t, exp.Save(store))
	}

	var checkpoints = []*project.Checkpoint{{
		ID:           "1ccccccccc",
		Created:      time.Now().UTC().Add(-1 * time.Minute),
		ExperimentID: experiments[0].ID,
		Metrics: map[string]*param.Value{
			"metric-1": param.Float(0.1),
			"metric-2": param.Int(2),
		},
		Step: 10,
	}, {
		ID:           "2ccccccccc",
		Created:      time.Now().UTC(),
		ExperimentID: experiments[0].ID,
		Metrics: map[string]*param.Value{
			"metric-1": param.Float(0.01),
			"metric-2": param.Int(2),
		},
		Step: 20,
	}, {
		ID:           "3ccccccccc",
		Created:      time.Now().UTC(),
		ExperimentID: experiments[0].ID,
		Metrics: map[string]*param.Value{
			"metric-1": param.Float(0.02),
			"metric-2": param.Int(2),
		},
		Step: 20,
	}, {
		ID:           "4ccccccccc",
		Created:      time.Now().UTC(),
		ExperimentID: experiments[1].ID,
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

func TestListOutputTableWithPrimaryMetricOnlyChangedParams(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "metric-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}, {
			Name: "metric-3",
			Goal: config.GoalMinimize,
		}},
	}

	store := createTestData(t, workingDir, conf)

	actual := capturer.CaptureStdout(func() {
		err = Experiments(store, FormatTable, false, new(param.Filters), &param.Sorter{Key: "started"})
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     PARAM-1  LATEST CHECKPOINT  LABEL-1  LABEL-3  BEST CHECKPOINT    LABEL-1  LABEL-3
3eeeeee     2 minutes ago       stopped  10.1.1.2  ben      200
2eeeeee     about a minute ago  stopped  10.1.1.2  andreas  200      4cccccc (step 5)            0.5
1eeeeee     about a second ago  running  10.1.1.1  andreas  100      3cccccc (step 20)  0.02              2cccccc (step 20)  0.01
`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputTableWithPrimaryMetricAllParams(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "metric-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}, {
			Name: "metric-3",
			Goal: config.GoalMinimize,
		}},
	}

	store := createTestData(t, workingDir, conf)

	actual := capturer.CaptureStdout(func() {
		err = Experiments(store, FormatTable, true, new(param.Filters), &param.Sorter{Key: "started"})
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     PARAM-1  PARAM-2  PARAM-3  LATEST CHECKPOINT  LABEL-1  LABEL-3  BEST CHECKPOINT    LABEL-1  LABEL-3
3eeeeee     2 minutes ago       stopped  10.1.1.2  ben      200      hello    hi
2eeeeee     about a minute ago  stopped  10.1.1.2  andreas  200      hello    hi       4cccccc (step 5)            0.5
1eeeeee     about a second ago  running  10.1.1.1  andreas  100      hello             3cccccc (step 20)  0.02              2cccccc (step 20)  0.01
`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputTableFilter(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "metric-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}, {
			Name: "metric-3",
			Goal: config.GoalMinimize,
		}},
	}

	store := createTestData(t, workingDir, conf)
	filters, err := param.MakeFilters([]string{"step >= 5"})
	require.NoError(t, err)
	sorter := param.NewSorter("started")

	actual := capturer.CaptureStdout(func() {
		err = Experiments(store, FormatTable, false, filters, sorter)
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     PARAM-1  LATEST CHECKPOINT  LABEL-1  LABEL-3  BEST CHECKPOINT    LABEL-1  LABEL-3
2eeeeee     about a minute ago  stopped  10.1.1.2  andreas  200      4cccccc (step 5)            0.5
1eeeeee     about a second ago  running  10.1.1.1  andreas  100      3cccccc (step 20)  0.02              2cccccc (step 20)  0.01
`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputTableFilterRunning(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "metric-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}, {
			Name: "metric-3",
			Goal: config.GoalMinimize,
		}},
	}

	store := createTestData(t, workingDir, conf)
	filters, err := param.MakeFilters([]string{"status = running"})
	require.NoError(t, err)
	sorter := param.NewSorter("started")

	actual := capturer.CaptureStdout(func() {
		err = Experiments(store, FormatTable, false, filters, sorter)
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     LATEST CHECKPOINT  LABEL-1  BEST CHECKPOINT    LABEL-1
1eeeeee     about a second ago  running  10.1.1.1  andreas  3cccccc (step 20)  0.02     2cccccc (step 20)  0.01
`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputTableSort(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "metric-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}, {
			Name: "metric-3",
			Goal: config.GoalMinimize,
		}},
	}

	store := createTestData(t, workingDir, conf)
	sorter := param.NewSorter("started-desc")

	actual := capturer.CaptureStdout(func() {
		err = Experiments(store, FormatTable, false, new(param.Filters), sorter)
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     PARAM-1  LATEST CHECKPOINT  LABEL-1  LABEL-3  BEST CHECKPOINT    LABEL-1  LABEL-3
1eeeeee     about a second ago  running  10.1.1.1  andreas  100      3cccccc (step 20)  0.02              2cccccc (step 20)  0.01
2eeeeee     about a minute ago  stopped  10.1.1.2  andreas  200      4cccccc (step 5)            0.5
3eeeeee     2 minutes ago       stopped  10.1.1.2  ben      200
`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListJSON(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "replicate-test")
	require.NoError(t, err)
	storageDir := path.Join(workingDir, ".replicate/storage")

	storage, err := storage.NewDiskStorage(storageDir)
	require.NoError(t, err)
	defer os.RemoveAll(storageDir)

	// Experiment no longer running
	exp := &project.Experiment{
		ID:      hash.Random(),
		Created: time.Now().UTC(),
		Params: map[string]*param.Value{
			"learning_rate": param.Float(0.001),
		},
		Command: "train.py --gamma 1.2",
	}
	require.NoError(t, exp.Save(storage))
	require.NoError(t, err)
	require.NoError(t, project.CreateHeartbeat(storage, exp.ID, time.Now().UTC().Add(-24*time.Hour)))
	com := project.NewCheckpoint(exp.ID, map[string]*param.Value{
		"accuracy": param.Float(0.987),
	})
	require.NoError(t, com.Save(storage, workingDir))

	// Experiment still running
	exp = &project.Experiment{
		ID:      hash.Random(),
		Created: time.Now().UTC(),
		Params: map[string]*param.Value{
			"learning_rate": param.Float(0.002),
		},
		Command: "train.py --gamma 1.5",
	}
	require.NoError(t, exp.Save(storage))
	require.NoError(t, err)
	require.NoError(t, project.CreateHeartbeat(storage, exp.ID, time.Now().UTC()))
	com = project.NewCheckpoint(exp.ID, map[string]*param.Value{
		"accuracy": param.Float(0.987),
	})
	require.NoError(t, com.Save(storage, workingDir))

	// replicate ls
	actual := capturer.CaptureStdout(func() {
		err = Experiments(storage, FormatJSON, true, new(param.Filters), &param.Sorter{Key: "started"})
	})
	require.NoError(t, err)

	experiments := make([]ListExperiment, 0)
	require.NoError(t, json.Unmarshal([]byte(actual), &experiments))
	require.Equal(t, 2, len(experiments))

	require.Equal(t, param.Float(0.001), experiments[0].Params["learning_rate"])
	require.Equal(t, "train.py --gamma 1.2", experiments[0].Command)
	require.Equal(t, 1, experiments[0].NumCheckpoints)
	require.Equal(t, param.Float(0.987), experiments[0].LatestCheckpoint.Metrics["accuracy"])
	require.Equal(t, false, experiments[0].Running)

	require.Equal(t, param.Float(0.002), experiments[1].Params["learning_rate"])
	require.Equal(t, "train.py --gamma 1.5", experiments[1].Command)
	require.Equal(t, 1, experiments[1].NumCheckpoints)
	require.Equal(t, param.Float(0.987), experiments[1].LatestCheckpoint.Metrics["accuracy"])
	require.Equal(t, true, experiments[1].Running)
}
