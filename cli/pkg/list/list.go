package list

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/xeonx/timeago"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/project"
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
	LatestCommit *project.Commit         `json:"latest_commit"`
	BestCommit   *project.Commit         `json:"best_commit"`
	User         string                  `json:"user"`
	Host         string                  `json:"host"`
	Running      bool                    `json:"running"`

	// exclude config from json output
	Config *config.Config `json:"-"`
}

func (e *ListExperiment) LatestActivity() time.Time {
	if e.LatestCommit != nil {
		return e.LatestCommit.Created
	}
	return e.Created
}

func RunningExperiments(store storage.Storage, format string, allParams bool) error {
	proj := project.NewProject(store)
	listExperiments, err := createListExperiments(proj)
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
		return outputTable(running, allParams)
	}
	return fmt.Errorf("Unknown format: %s", format)
}

func Experiments(store storage.Storage, format string, allParams bool) error {
	proj := project.NewProject(store)
	listExperiments, err := createListExperiments(proj)
	if err != nil {
		return err
	}
	switch format {
	case FormatJSON:
		return outputJSON(listExperiments)
	case FormatTable:
		return outputTable(listExperiments, allParams)
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
func outputTable(experiments []*ListExperiment, allParams bool) error {
	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	expHeadings := getExperimentHeadings(experiments, !allParams)
	commitHeadings := getCommitHeadings(experiments)

	// does any experiment have a primary metric?
	hasPrimaryMetric := false
	for _, exp := range experiments {
		if exp.Config != nil && exp.Config.PrimaryMetric() != nil {
			hasPrimaryMetric = true
		}
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	keys := []string{"EXPERIMENT", "STARTED", "STATUS", "HOST", "USER"}
	keys = append(keys, upper(expHeadings)...)
	keys = append(keys, "LATEST", "STEP")
	keys = append(keys, upper(commitHeadings)...)
	if hasPrimaryMetric {
		keys = append(keys, "BEST", "STEP")
		keys = append(keys, upper(commitHeadings)...)
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

		latestCommitID := ""
		latestCommitStep := ""
		if exp.LatestCommit != nil {
			latestCommitID = exp.LatestCommit.ID[:7]
			latestCommitStep = strconv.Itoa(exp.LatestCommit.Step)
		}
		fmt.Fprintf(tw, "%s\t", latestCommitID)
		fmt.Fprintf(tw, "%s\t", latestCommitStep)

		// latest commit labels
		for _, heading := range commitHeadings {
			val := ""
			if exp.LatestCommit != nil {
				if v, ok := exp.LatestCommit.Labels[heading]; ok {
					val = v.String()
				}
			}
			fmt.Fprintf(tw, "%s\t", val)
		}

		bestCommitID := ""
		bestStepID := ""

		if exp.BestCommit != nil {
			bestCommitID = exp.BestCommit.ID[:7]
			bestStepID = strconv.Itoa(exp.BestCommit.Step)
		}
		fmt.Fprintf(tw, "%s\t", bestCommitID)
		fmt.Fprintf(tw, "%s\t", bestStepID)

		// best commit labels
		for _, heading := range commitHeadings {
			val := ""
			if exp.BestCommit != nil {
				if v, ok := exp.BestCommit.Labels[heading]; ok {
					val = v.String()
				}
			}
			fmt.Fprintf(tw, "%s\t", val)
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
func getCommitHeadings(experiments []*ListExperiment) []string {
	metricNameSet := map[string]bool{}
	commitHeadingSet := map[string]bool{}

	for _, exp := range experiments {
		if exp.Config == nil {
			continue
		}
		for _, metric := range exp.Config.Metrics {
			metricNameSet[metric.Name] = true
		}
	}
	for _, exp := range experiments {
		if exp.LatestCommit != nil {
			for key := range exp.LatestCommit.Labels {
				if _, ok := metricNameSet[key]; ok {
					commitHeadingSet[key] = true
				}
			}
		}
	}

	return slices.StringKeys(commitHeadingSet)
}

func createListExperiments(proj *project.Project) ([]*ListExperiment, error) {
	experiments, err := proj.Experiments()
	if err != nil {
		return nil, err
	}
	ret := []*ListExperiment{}
	for _, exp := range experiments {
		listExperiment := ListExperiment{
			ID:      exp.ID,
			Params:  exp.Params,
			Created: exp.Created,
			Host:    exp.Host,
			User:    exp.User,
			Config:  exp.Config,
		}

		commits, err := proj.ExperimentCommits(exp.ID)
		if err != nil {
			return nil, err
		}
		latest, err := proj.ExperimentLatestCommit(exp.ID)
		if err != nil {
			return nil, err
		}
		best, err := proj.ExperimentBestCommit(exp.ID)
		if err != nil {
			return nil, err
		}
		running, err := proj.ExperimentIsRunning(exp.ID)
		if err != nil {
			return nil, err
		}
		listExperiment.LatestCommit = latest
		listExperiment.BestCommit = best
		listExperiment.NumCommits = len(commits)
		listExperiment.Running = running

		ret = append(ret, &listExperiment)

	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].LatestActivity().Before(ret[j].LatestActivity())
	})

	return ret, nil

}

func upper(in []string) []string {
	ret := make([]string, len(in))
	for i, s := range in {
		ret[i] = strings.ToUpper(s)
	}
	return ret
}
