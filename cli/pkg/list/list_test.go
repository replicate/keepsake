package list

import (
	"strings"
	"testing"
	"time"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/require"

	"replicate.ai/cli/pkg/commit"
	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/param"
)

var experiments = []*experiment.Experiment{{
	ID:      "1eeeeeeeee",
	Created: time.Now().Add(-10 * time.Second),
	Params: map[string]*param.Value{
		"param-1": param.Int(100),
		"param-2": param.String("hello"),
	},
	Host:    "10.1.1.1",
	User:    "andreas",
	Running: true,
}, {
	ID:      "2eeeeeeeee",
	Created: time.Now(),
	Params: map[string]*param.Value{
		"param-1": param.Int(200),
		"param-2": param.String("hello"),
		"param-3": param.String("hi"),
	},
	Host:    "10.1.1.2",
	User:    "andreas",
	Running: false,
}}
var commits = []*commit.Commit{{
	ID:         "1ccccccccc",
	Created:    time.Now().Add(-10 * time.Second),
	Experiment: experiments[0],
	Labels: map[string]*param.Value{
		"label-1": param.Float(0.1),
		"label-2": param.Int(2),
	},
	Step: 10,
}, {
	ID:         "2ccccccccc",
	Created:    time.Now(),
	Experiment: experiments[0],
	Labels: map[string]*param.Value{
		"label-1": param.Float(0.01),
		"label-2": param.Int(2),
	},
	Step: 20,
}, {
	ID:         "3ccccccccc",
	Created:    time.Now(),
	Experiment: experiments[0],
	Labels: map[string]*param.Value{
		"label-1": param.Float(0.02),
		"label-2": param.Int(2),
	},
	Step: 20,
}, {
	ID:         "4ccccccccc",
	Created:    time.Now(),
	Experiment: experiments[1],
	Labels: map[string]*param.Value{
		"label-3": param.Float(0.5),
	},
	Step: 5,
}}

func TestGroupCommitsWithoutPrimaryMetric(t *testing.T) {
	conf := &config.Config{
		Metrics: []config.Metric{{
			Name: "label-1",
			Goal: config.GoalMinimize,
		}},
	}
	expected := []*GroupedExperiment{{
		ID:      "1eeeeeeeee",
		Created: experiments[0].Created,
		Params: map[string]*param.Value{
			"param-1": param.Int(100),
			"param-2": param.String("hello"),
		},
		NumCommits:   3,
		LatestCommit: commits[2],
		BestCommit:   nil,
		Host:         "10.1.1.1",
		User:         "andreas",
		Running:      true,
	}, {
		ID:      "2eeeeeeeee",
		Created: experiments[1].Created,
		Params: map[string]*param.Value{
			"param-1": param.Int(200),
			"param-2": param.String("hello"),
			"param-3": param.String("hi"),
		},
		NumCommits:   1,
		LatestCommit: commits[3],
		BestCommit:   nil,
		Host:         "10.1.1.2",
		User:         "andreas",
		Running:      false,
	}}

	actual := groupCommits(conf, commits)
	require.Equal(t, expected, actual)
}

func TestGroupCommitsWithPrimaryMetric(t *testing.T) {
	conf := &config.Config{
		Metrics: []config.Metric{{
			Name:    "label-1",
			Goal:    config.GoalMinimize,
			Primary: true,
		}},
	}
	expected := []*GroupedExperiment{{
		ID:      "1eeeeeeeee",
		Created: experiments[0].Created,
		Params: map[string]*param.Value{
			"param-1": param.Int(100),
			"param-2": param.String("hello"),
		},
		NumCommits:   3,
		LatestCommit: commits[2],
		BestCommit:   commits[1],
		Host:         "10.1.1.1",
		User:         "andreas",
		Running:      true,
	}, {
		ID:      "2eeeeeeeee",
		Created: experiments[1].Created,
		Params: map[string]*param.Value{
			"param-1": param.Int(200),
			"param-2": param.String("hello"),
			"param-3": param.String("hi"),
		},
		NumCommits:   1,
		LatestCommit: commits[3],
		BestCommit:   nil,
		Host:         "10.1.1.2",
		User:         "andreas",
		Running:      false,
	}}

	actual := groupCommits(conf, commits)
	require.Equal(t, expected, actual)
}

func TestOutputTableWithPrimaryMetricOnlyChangedParams(t *testing.T) {
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
	experiments := groupCommits(conf, commits)
	actual := capturer.CaptureStdout(func() {
		err := outputTable(conf, experiments, false)
		require.NoError(t, err)
	})
	expected := `
experiment  started             status   host      user     param-1  latest   step  label-1  label-3  best     step  label-1  label-3
1eeeeee     10 seconds ago      running  10.1.1.1  andreas  100      3cccccc  20    0.02              2cccccc  20    0.01
2eeeeee     about a second ago  stopped  10.1.1.2  andreas  200      4cccccc  5              0.5      N/A
`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = trimRightLines(actual)
	require.Equal(t, expected, actual)
}

func TestOutputTableWithPrimaryMetricAllParams(t *testing.T) {
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
	experiments := groupCommits(conf, commits)
	actual := capturer.CaptureStdout(func() {
		err := outputTable(conf, experiments, true)
		require.NoError(t, err)
	})
	expected := `
experiment  started             status   host      user     param-1  param-2  param-3  latest   step  label-1  label-3  best     step  label-1  label-3
1eeeeee     10 seconds ago      running  10.1.1.1  andreas  100      hello             3cccccc  20    0.02              2cccccc  20    0.01
2eeeeee     about a second ago  stopped  10.1.1.2  andreas  200      hello    hi       4cccccc  5              0.5      N/A
`
	expected = expected[1:] // strip initial whitespace, added for readability
	actual = trimRightLines(actual)
	require.Equal(t, expected, actual)
}

func trimRightLines(s string) string {
	lines := []string{}
	for _, line := range strings.Split(s, "\n") {
		lines = append(lines, strings.TrimRight(line, " "))
	}
	return strings.Join(lines, "\n")
}
