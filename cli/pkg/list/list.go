package list

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/xeonx/timeago"

	"replicate.ai/cli/pkg/commit"
	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/slices"
	"replicate.ai/cli/pkg/storage"
)

const FormatJSON = "json"
const FormatTable = "table"

type Metric struct {
	Primary bool
	Name    string
	Value   float64
}

type ListExperiment struct {
	ID           string                  `json:"id"`
	Created      time.Time               `json:"created"`
	Params       map[string]*param.Value `json:"params"`
	NumCommits   int                     `json:"num_commits"`
	LatestCommit *commit.Commit          `json:"latest_commit"`
	BestCommit   *commit.Commit          `json:"best_commit"`
	User         string                  `json:"user"`
	Host         string                  `json:"host"`
	Running      bool                    `json:"running"`
}

func RunningExperiments(store storage.Storage, format string, allParams bool) error {
	experiments, err := experiment.List(store)
	if err != nil {
		return err
	}
	conf := configFromExperiments(experiments)

	listExperiments, err := createListExperiments(conf, store, experiments)
	if err != nil {
		return err
	}

	running := []*ListExperiment{}
	for _, exp := range listExperiments {
		if exp.Running {
			running = append(running, exp)
		}
	}

	switch format {
	case FormatJSON:
		return outputJSON(running)
	case FormatTable:
		return outputTable(conf, running, allParams)
	}
	return fmt.Errorf("Unknown format: %s", format)
}

func Experiments(store storage.Storage, format string, allParams bool) error {
	experiments, err := experiment.List(store)
	if err != nil {
		return err
	}

	// TODO(andreas): this means we read config from storage every
	// time. we should use local config if it exists, so you can
	// update metrics etc.
	// See also discussion in https://github.com/replicate/replicate/pull/44
	conf := configFromExperiments(experiments)

	listExperiments, err := createListExperiments(conf, store, experiments)
	if err != nil {
		return err
	}

	switch format {
	case FormatJSON:
		return outputJSON(listExperiments)
	case FormatTable:
		return outputTable(conf, listExperiments, allParams)
	}
	return fmt.Errorf("Unknown format: %s", format)
}

func outputJSON(experiments []*ListExperiment) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(experiments)
}

// output something like (TODO: this is getting very wide)
// experiment  started             status   host      user     param-1  latest   step  label-1  best     step  label-1
// 1eeeeee     10 seconds ago      running  10.1.1.1  andreas  100      3cccccc  20    0.02     2cccccc  20    0.01
// 2eeeeee     about a second ago  stopped  10.1.1.2  andreas  200      4cccccc  5              N/A
func outputTable(conf *config.Config, experiments []*ListExperiment, allParams bool) error {
	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	expHeadings := getExperimentHeadings(experiments, !allParams)
	commitHeadings := getCommitHeadings(conf, experiments)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	keys := []string{"experiment", "started", "status", "host", "user"}
	keys = append(keys, expHeadings...)
	keys = append(keys, "latest", "step")
	keys = append(keys, commitHeadings...)
	if conf.PrimaryMetric() != nil {
		keys = append(keys, "best", "step")
		keys = append(keys, commitHeadings...)
	}

	for i, key := range keys {
		fmt.Fprintf(tw, "%s", key)
		if i < len(keys)-1 {
			fmt.Fprint(tw, "\t")
		}
	}
	fmt.Fprint(tw, "\n")

	for _, exp := range experiments {
		// experiment
		fmt.Fprintf(tw, "%s\t", exp.ID[:7])

		// started
		fmt.Fprintf(tw, "%s\t", formatTime(exp.Created))

		// status
		if exp.Running {
			fmt.Fprint(tw, "running\t")
		} else {
			fmt.Fprint(tw, "stopped\t")
		}

		// host
		fmt.Fprintf(tw, "%s\t", exp.Host)

		// user
		fmt.Fprintf(tw, "%s\t", exp.User)

		// experiment params
		for _, heading := range expHeadings {
			if val, ok := exp.Params[heading]; ok {
				fmt.Fprintf(tw, "%v", val)
			}
			fmt.Fprintf(tw, "\t")
		}

		// latest commit id
		fmt.Fprintf(tw, "%s\t", exp.LatestCommit.ID[:7])

		// latest step
		fmt.Fprintf(tw, "%d\t", exp.LatestCommit.Step)

		// latest commit labels
		for _, heading := range commitHeadings {
			if val, ok := exp.LatestCommit.Labels[heading]; ok {
				fmt.Fprintf(tw, "%v", val)
			}
			fmt.Fprintf(tw, "\t")
		}

		if exp.BestCommit != nil {
			// best commit id
			fmt.Fprintf(tw, "%s\t", exp.BestCommit.ID[:7])

			// best step
			fmt.Fprintf(tw, "%d\t", exp.BestCommit.Step)

			// best commit labels
			for _, heading := range commitHeadings {
				if val, ok := exp.BestCommit.Labels[heading]; ok {
					fmt.Fprintf(tw, "%v", val)
				}
				fmt.Fprintf(tw, "\t")
			}
		} else if conf.PrimaryMetric() != nil {
			fmt.Fprintf(tw, "N/A")
		}

		// newline!
		fmt.Fprint(tw, "\n")
	}

	if err := tw.Flush(); err != nil {
		return err
	}

	return nil
}

func formatTime(t time.Time) string {
	return timeago.English.Format(t)
}

// get experiment params. if onlyChangedParams is true, only return
// params which have changed across experiments
func getExperimentHeadings(experiments []*ListExperiment, onlyChangedParams bool) []string {
	expHeadingSet := map[string]bool{}

	if onlyChangedParams {
		paramValues := map[string]*param.Value{}
		for _, exp := range experiments {
			for key, val := range exp.Params {
				firstVal, ok := paramValues[key]
				if ok {
					notEqual, err := firstVal.NotEqual(val)
					if err != nil {
						console.Warn("%s", err)
					} else if notEqual {
						expHeadingSet[key] = true
					}
				} else {
					paramValues[key] = val
				}
			}
		}
	} else {
		for _, exp := range experiments {
			for key := range exp.Params {
				expHeadingSet[key] = true
			}
		}
	}

	return slices.StringKeys(expHeadingSet)
}

// get commit labels that are also defined as metrics in config
func getCommitHeadings(conf *config.Config, experiments []*ListExperiment) []string {
	metricNameSet := map[string]bool{}
	commitHeadingSet := map[string]bool{}

	for _, metric := range conf.Metrics {
		metricNameSet[metric.Name] = true
	}
	for _, exp := range experiments {
		for key := range exp.LatestCommit.Labels {
			if _, ok := metricNameSet[key]; ok {
				commitHeadingSet[key] = true
			}
		}
	}

	return slices.StringKeys(commitHeadingSet)
}

func createListExperiments(conf *config.Config, store storage.Storage, experiments []*experiment.Experiment) ([]*ListExperiment, error) {
	commits, err := commit.ListCommits(store)
	if err != nil {
		return nil, err
	}
	expIDToCommits := make(map[string][]*commit.Commit)
	for _, com := range commits {
		expID := com.ExperimentID
		if _, ok := expIDToCommits[expID]; !ok {
			expIDToCommits[expID] = []*commit.Commit{}
		}
		expIDToCommits[expID] = append(expIDToCommits[expID], com)
	}

	ret := []*ListExperiment{}

	for _, exp := range experiments {
		listExperiment := ListExperiment{
			ID:      exp.ID,
			Params:  exp.Params,
			Created: exp.Created,
			Host:    exp.Host,
			User:    exp.User,
		}

		commits, ok := expIDToCommits[exp.ID]
		if ok {
			listExperiment.LatestCommit = getLatestCommit(commits)
			listExperiment.BestCommit = getBestCommit(conf, commits)
			listExperiment.NumCommits = len(commits)
		}

		heartbeat, err := experiment.LoadHeartbeat(store, exp.ID)
		// TODO: handle errors other than heartbeat not existing. requires some standard not exist error for storage, and maybe heartbeats too.
		if err == nil {
			listExperiment.Running = heartbeat.IsRunning()
		}

		ret = append(ret, &listExperiment)

	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].LatestCommit.Created.Before(ret[j].LatestCommit.Created)
	})

	return ret, nil

}

func getLatestCommit(commits []*commit.Commit) *commit.Commit {
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Created.Before(commits[j].Created)
	})
	return commits[len(commits)-1]
}

// return the best commit according to the primary metric, or nil
// if primary metric is not defined or if none of the commits have
// the primary metric defined
func getBestCommit(conf *config.Config, commits []*commit.Commit) *commit.Commit {
	primaryMetric := conf.PrimaryMetric()
	if primaryMetric == nil {
		return nil
	}

	// sort commits in ascending order, according to the primary metric
	sort.Slice(commits, func(i, j int) bool {
		iVal, iOK := commits[i].Labels[primaryMetric.Name]
		jVal, jOK := commits[j].Labels[primaryMetric.Name]
		if !iOK {
			return true
		}
		if !jOK {
			return false
		}
		if primaryMetric.Goal == config.GoalMaximize {
			less, err := iVal.LessThan(jVal)
			if err != nil {
				console.Warn("Got error when comparing metrics: %s", err)
			}
			return less
		} else {
			greater, err := iVal.GreaterThan(jVal)
			if err != nil {
				console.Warn("Got error when comparing metrics: %s", err)
			}
			return greater
		}
	})
	best := commits[len(commits)-1]

	// if the last (best) commit in the sorted list doesn't have
	// a value for the primary metric, none of them do
	if _, ok := best.Labels[primaryMetric.Name]; !ok {
		return nil
	}

	return best
}

// pull out the saved config from the commits list
// TODO(andreas): what to do if config changes between experiments
// See also discussion in https://github.com/replicate/replicate/pull/44
func configFromExperiments(experiments []*experiment.Experiment) *config.Config {
	if len(experiments) == 0 || experiments[0].Config == nil {
		// FIXME (bfirsh): Should this be getDefaultConfig()? if so, we need to get the working dir somehow
		// Maybe config should just be nullable.
		return new(config.Config)
	}
	// FIXME (bfirsh): this isn't sorted, so is kind of meaningless. this implementation is also broken, and will
	// fix in subsequent commit.
	// see https://replicatehq.slack.com/archives/CPRGK33J5/p1596491607005100
	return experiments[0].Config
}
