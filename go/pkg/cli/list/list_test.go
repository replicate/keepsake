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

	"github.com/replicate/keepsake/go/pkg/config"
	"github.com/replicate/keepsake/go/pkg/hash"
	"github.com/replicate/keepsake/go/pkg/param"
	"github.com/replicate/keepsake/go/pkg/project"
	"github.com/replicate/keepsake/go/pkg/repository"
	"github.com/replicate/keepsake/go/pkg/testutil"
)

func createTestData(t *testing.T, workingDir string, conf *config.Config) repository.Repository {
	repo, err := repository.NewDiskRepository(path.Join(workingDir, ".keepsake"))
	require.NoError(t, err)

	require.NoError(t, err)
	var experiments = []*project.Experiment{{
		ID:      "1eeeeeeeee",
		Created: time.Now().UTC(),
		Params: param.ValueMap{
			"param-1": param.Int(100),
			"param-2": param.String("hello"),
		},
		Host:    "10.1.1.1",
		User:    "andreas",
		Config:  conf,
		Command: "train.py --foo bar",
		Checkpoints: []*project.Checkpoint{
			{
				ID:      "1ccccccccc",
				Created: time.Now().UTC().Add(-1 * time.Minute),
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
				Created: time.Now().UTC(),
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
				Created: time.Now().UTC(),
				Metrics: param.ValueMap{
					"metric-1": param.Float(0.02),
					"metric-2": param.Int(2),
					// test it works with None
					"metric-3": param.None(),
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
		Created: time.Now().UTC().Add(-1 * time.Minute),
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
				Created: time.Now().UTC(),
				Metrics: param.ValueMap{
					"metric-3": param.Float(0.5),
				},
				PrimaryMetric: &project.PrimaryMetric{
					Name: "metric-1",
					Goal: project.GoalMinimize,
				},
				Step: 5,
			},
		},
	}, {
		ID:      "3eeeeeeeee",
		Created: time.Now().UTC().Add(-2 * time.Minute),
		Params: param.ValueMap{
			"param-1": param.Int(200),
			"param-2": param.String("hello"),
			"param-3": param.String("hi"),
			// test it works with None
			"param-4": param.None(),
			"param-5": param.String("__verylongparameterstring__"),
		},
		Host:   "10.1.1.2",
		User:   "ben",
		Config: conf,
	}}
	for _, exp := range experiments {
		require.NoError(t, exp.Save(repo))
	}

	require.NoError(t, project.CreateHeartbeat(repo, experiments[0].ID, time.Now().UTC()))
	require.NoError(t, project.CreateHeartbeat(repo, experiments[1].ID, time.Now().UTC().Add(-1*time.Minute)))

	return repo
}

func TestListOutputTableWithPrimaryMetricOnlyChangedParams(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}

	repo := createTestData(t, workingDir, conf)

	actual := capturer.CaptureStdout(func() {
		err = Experiments(repo, FormatTable, false, new(param.Filters), &param.Sorter{Key: "started"})
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     PARAMS       BEST CHECKPOINT    LATEST CHECKPOINT
3eeeeee     2 minutes ago       stopped  10.1.1.2  ben      param-1=200

2eeeeee     about a minute ago  stopped  10.1.1.2  andreas  param-1=200                     4cccccc (step 5)

1eeeeee     about a second ago  running  10.1.1.1  andreas  param-1=100  2cccccc (step 20)  3cccccc (step 20)
                                                                         metric-1=0.01      metric-1=0.02

`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputTableWithPrimaryMetricAll(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createTestData(t, workingDir, conf)

	actual := capturer.CaptureStdout(func() {
		err = Experiments(repo, FormatTable, true, new(param.Filters), &param.Sorter{Key: "started"})
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     PARAMS                        BEST CHECKPOINT    LATEST CHECKPOINT
3eeeeee     2 minutes ago       stopped  10.1.1.2  ben      param-1=200
                                                            param-2=hello
                                                            param-3=hi
                                                            param-4=null
                                                            param-5=__verylongparamet...

2eeeeee     about a minute ago  stopped  10.1.1.2  andreas  param-1=200                                      4cccccc (step 5)
                                                            param-2=hello                                    metric-3=0.5
                                                            param-3=hi

1eeeeee     about a second ago  running  10.1.1.1  andreas  param-1=100                   2cccccc (step 20)  3cccccc (step 20)
                                                            param-2=hello                 metric-1=0.01      metric-1=0.02
                                                                                          metric-2=2         metric-2=2
                                                                                                             metric-3=null

`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputFullTableWithPrimaryMetricAll(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createTestData(t, workingDir, conf)

	actual := capturer.CaptureStdout(func() {
		err = Experiments(repo, FormatFullTable, true, new(param.Filters), &param.Sorter{Key: "started"})
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     PARAMS                               BEST CHECKPOINT    LATEST CHECKPOINT
3eeeeee     2 minutes ago       stopped  10.1.1.2  ben      param-1=200
                                                            param-2=hello
                                                            param-3=hi
                                                            param-4=null
                                                            param-5=__verylongparameterstring__

2eeeeee     about a minute ago  stopped  10.1.1.2  andreas  param-1=200                                             4cccccc (step 5)
                                                            param-2=hello                                           metric-3=0.5
                                                            param-3=hi

1eeeeee     about a second ago  running  10.1.1.1  andreas  param-1=100                          2cccccc (step 20)  3cccccc (step 20)
                                                            param-2=hello                        metric-1=0.01      metric-1=0.02
                                                                                                 metric-2=2         metric-2=2
                                                                                                                    metric-3=null

`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputTableFilter(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createTestData(t, workingDir, conf)
	filters, err := param.MakeFilters([]string{"step >= 5"})
	require.NoError(t, err)
	sorter := param.NewSorter("started")

	actual := capturer.CaptureStdout(func() {
		err = Experiments(repo, FormatTable, false, filters, sorter)
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      PARAMS       BEST CHECKPOINT    LATEST CHECKPOINT
2eeeeee     about a minute ago  stopped  10.1.1.2  param-1=200                     4cccccc (step 5)

1eeeeee     about a second ago  running  10.1.1.1  param-1=100  2cccccc (step 20)  3cccccc (step 20)
                                                                metric-1=0.01      metric-1=0.02

`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputTableFilterRunning(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createTestData(t, workingDir, conf)
	filters, err := param.MakeFilters([]string{"status = running"})
	require.NoError(t, err)
	sorter := param.NewSorter("started")

	actual := capturer.CaptureStdout(func() {
		err = Experiments(repo, FormatTable, false, filters, sorter)
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   PARAMS  BEST CHECKPOINT    LATEST CHECKPOINT
1eeeeee     about a second ago  running          2cccccc (step 20)  3cccccc (step 20)
                                                 metric-1=0.01      metric-1=0.02

`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListOutputTableSort(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	defer os.RemoveAll(workingDir)

	conf := &config.Config{}
	repo := createTestData(t, workingDir, conf)
	sorter := param.NewSorter("started-desc")

	actual := capturer.CaptureStdout(func() {
		err = Experiments(repo, FormatTable, false, new(param.Filters), sorter)
	})
	require.NoError(t, err)
	expected := `
EXPERIMENT  STARTED             STATUS   HOST      USER     PARAMS       BEST CHECKPOINT    LATEST CHECKPOINT
1eeeeee     about a second ago  running  10.1.1.1  andreas  param-1=100  2cccccc (step 20)  3cccccc (step 20)
                                                                         metric-1=0.01      metric-1=0.02

2eeeeee     about a minute ago  stopped  10.1.1.2  andreas  param-1=200                     4cccccc (step 5)

3eeeeee     2 minutes ago       stopped  10.1.1.2  ben      param-1=200

`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = testutil.TrimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestListJSON(t *testing.T) {
	workingDir, err := ioutil.TempDir("", "keepsake-test")
	require.NoError(t, err)
	repoDir := path.Join(workingDir, ".keepsake")

	repository, err := repository.NewDiskRepository(repoDir)
	require.NoError(t, err)
	defer os.RemoveAll(repoDir)

	// Experiment no longer running
	exp := &project.Experiment{
		ID:      hash.Random(),
		Created: time.Now().UTC(),
		Params: param.ValueMap{
			"learning_rate": param.Float(0.001),
		},
		Command: "train.py --gamma 1.2",
		Checkpoints: []*project.Checkpoint{
			project.NewCheckpoint(param.ValueMap{
				"accuracy": param.Float(0.987),
			}),
		},
	}
	require.NoError(t, exp.Save(repository))
	require.NoError(t, err)
	require.NoError(t, project.CreateHeartbeat(repository, exp.ID, time.Now().UTC().Add(-24*time.Hour)))

	// Experiment still running
	exp = &project.Experiment{
		ID:      hash.Random(),
		Created: time.Now().UTC(),
		Params: param.ValueMap{
			"learning_rate": param.Float(0.002),
		},
		Command: "train.py --gamma 1.5",
		Checkpoints: []*project.Checkpoint{
			project.NewCheckpoint(param.ValueMap{
				"accuracy": param.Float(0.987),
			}),
		},
	}
	require.NoError(t, exp.Save(repository))
	require.NoError(t, err)
	require.NoError(t, project.CreateHeartbeat(repository, exp.ID, time.Now().UTC()))

	// keepsake ls
	actual := capturer.CaptureStdout(func() {
		err = Experiments(repository, FormatJSON, true, new(param.Filters), &param.Sorter{Key: "started"})
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
