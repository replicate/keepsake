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

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/slices"
	"replicate.ai/cli/pkg/storage"
)

type Format int

const (
	FormatJSON = iota
	FormatTable
)

const valueMaxLength = 10
const valueTruncate = 5

type Metric struct {
	Primary bool
	Name    string
	Value   float64
}

type ListExperiment struct {
	ID           string                  `json:"id"`
	Created      time.Time               `json:"created"`
	Params       map[string]*param.Value `json:"params"`
	Command      string                  `json:"command"`
	NumCommits   int                     `json:"num_commits"`
	LatestCommit *project.Commit         `json:"latest_commit"`
	BestCommit   *project.Commit         `json:"best_commit"`
	User         string                  `json:"user"`
	Host         string                  `json:"host"`
	Running      bool                    `json:"running"`

	// exclude config from json output
	Config *config.Config `json:"-"`
}

// TODO(andreas): make this safer and validate user inputs against these strings
func (exp *ListExperiment) GetValue(name string) *param.Value {
	if name == "started" {
		// floating point timestamp used in sorting
		return param.Float(float64(exp.Created.Unix()))
	}
	if name == "step" {
		if exp.LatestCommit != nil {
			return param.Int(exp.LatestCommit.Step)
		}
		return param.Int(0)
	}
	if name == "user" {
		return param.String(exp.User)
	}
	if name == "host" {
		return param.String(exp.Host)
	}
	if name == "command" {
		return param.String(exp.Command)
	}
	if name == "status" {
		if exp.Running {
			return param.String("running")
		}
		return param.String("stopped")
	}
	if exp.BestCommit != nil {
		if val, ok := exp.BestCommit.Labels[name]; ok {
			return val
		}
	}
	if val, ok := exp.Params[name]; ok {
		return val
	}
	return nil
}

func Experiments(store storage.Storage, format Format, allParams bool, filters *param.Filters, sorter *param.Sorter) error {
	proj := project.NewProject(store)
	if storage.NeedsCaching(store) {
		console.Info("Fetching experiments from %q...", store.RootURL())
	}
	listExperiments, err := createListExperiments(proj, filters)
	if err != nil {
		return err
	}
	sort.Slice(listExperiments, func(i, j int) bool {
		return sorter.LessThan(listExperiments[i], listExperiments[j])
	})

	switch format {
	case FormatJSON:
		return outputJSON(listExperiments)
	case FormatTable:
		return outputTable(listExperiments, allParams)
	}
	panic(fmt.Sprintf("Unknown format: %d", format))
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
	keys = append(keys, "LATEST COMMIT")
	keys = append(keys, upper(commitHeadings)...)
	if hasPrimaryMetric {
		keys = append(keys, "BEST COMMIT")
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
		fmt.Fprintf(tw, "%s\t", console.FormatTime(exp.Created))

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

		latestCommit := ""
		if exp.LatestCommit != nil {
			latestCommit = fmt.Sprintf("%s (step %s)", exp.LatestCommit.ShortID(), strconv.Itoa(exp.LatestCommit.Step))
		}
		fmt.Fprintf(tw, "%s\t", latestCommit)

		// latest commit labels
		for _, heading := range commitHeadings {
			val := ""
			if exp.LatestCommit != nil {
				if v, ok := exp.LatestCommit.Labels[heading]; ok {
					val = v.ShortString(valueMaxLength, valueTruncate)
				}
			}
			fmt.Fprintf(tw, "%s\t", val)
		}

		bestCommit := ""

		if exp.BestCommit != nil {
			bestCommit = fmt.Sprintf("%s (step %s)", exp.BestCommit.ShortID(), strconv.Itoa(exp.BestCommit.Step))
		}
		fmt.Fprintf(tw, "%s\t", bestCommit)

		// best commit labels
		for _, heading := range commitHeadings {
			val := ""
			if exp.BestCommit != nil {
				if v, ok := exp.BestCommit.Labels[heading]; ok {
					val = v.ShortString(valueMaxLength, valueTruncate)
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

func createListExperiments(proj *project.Project, filters *param.Filters) ([]*ListExperiment, error) {
	experiments, err := proj.Experiments()
	if err != nil {
		return nil, err
	}
	ret := []*ListExperiment{}
	for _, exp := range experiments {
		listExperiment := &ListExperiment{
			ID:      exp.ID,
			Params:  exp.Params,
			Command: exp.Command,
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

		match, err := filters.Matches(listExperiment)
		if err != nil {
			return nil, err
		}
		if !match {
			continue
		}
		ret = append(ret, listExperiment)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Created.Before(ret[j].Created)
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
