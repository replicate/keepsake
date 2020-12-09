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

	"github.com/replicate/replicate/go/pkg/config"
	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/param"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/repository"
	"github.com/replicate/replicate/go/pkg/slices"
)

type Format int

const (
	FormatJSON = iota
	FormatTable
	FormatQuiet
)

const valueMaxLength = 20
const valueTruncate = 5

type ListExperiment struct {
	ID               string              `json:"id"`
	Created          time.Time           `json:"created"`
	Params           param.ValueMap      `json:"params"`
	Command          string              `json:"command"`
	NumCheckpoints   int                 `json:"num_checkpoints"`
	LatestCheckpoint *project.Checkpoint `json:"latest_checkpoint"`
	BestCheckpoint   *project.Checkpoint `json:"best_checkpoint"`
	User             string              `json:"user"`
	Host             string              `json:"host"`
	Running          bool                `json:"running"`

	// exclude config from json output
	Config *config.Config `json:"-"`
}

// We should add some validation and better error messages, see https://github.com/replicate/replicate/issues/340
func (exp *ListExperiment) GetValue(name string) param.Value {
	if name == "started" {
		// floating point timestamp used in sorting
		return param.Float(float64(exp.Created.Unix()))
	}
	if name == "step" {
		if exp.LatestCheckpoint != nil {
			return param.Int(int64(exp.LatestCheckpoint.Step))
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
	if exp.BestCheckpoint != nil {
		if val, ok := exp.BestCheckpoint.Metrics[name]; ok {
			return val
		}
	}
	if val, ok := exp.Params[name]; ok {
		return val
	}
	return param.None()
}

func Experiments(repo repository.Repository, format Format, all bool, filters *param.Filters, sorter *param.Sorter) error {
	proj := project.NewProject(repo, "")
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
		return outputTable(listExperiments, all)
	case FormatQuiet:
		return outputQuiet(listExperiments)
	}
	panic(fmt.Sprintf("Unknown format: %d", format))
}

func outputQuiet(experiments []*ListExperiment) error {
	for _, exp := range experiments {
		fmt.Println(exp.ID)
	}
	return nil
}

func outputJSON(experiments []*ListExperiment) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(experiments)
}

// output something like (TODO: this is getting very wide)
// experiment  started             status   host      user     param-1  latest   step  metric-1  best     step  metric-1
// 1eeeeee     10 seconds ago      running  10.1.1.1  andreas  100      3cccccc  20    0.02     2cccccc  20    0.01
// 2eeeeee     about a second ago  stopped  10.1.1.2  andreas  200      4cccccc  5              N/A
func outputTable(experiments []*ListExperiment, all bool) error {
	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	paramsToDisplay := getParamsToDisplay(experiments, all)
	metricsToDisplay := getMetricsToDisplay(experiments, all)

	// does any experiment have a primary metric?
	hasBestCheckpoint := false
	for _, exp := range experiments {
		if exp.BestCheckpoint != nil {
			hasBestCheckpoint = true
			break
		}
	}

	// Hide various fields if they are all the same
	displayHost := false
	displayUser := false
	prevExp := experiments[0]
	for _, exp := range experiments {
		if exp.Host != prevExp.Host {
			displayHost = true
		}
		if exp.User != prevExp.User {
			displayUser = true
		}
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	keys := []string{"EXPERIMENT", "STARTED", "STATUS"}
	if displayHost {
		keys = append(keys, "HOST")
	}
	if displayUser {
		keys = append(keys, "USER")
	}
	keys = append(keys, upper(paramsToDisplay)...)
	keys = append(keys, "LATEST CHECKPOINT")
	keys = append(keys, upper(metricsToDisplay)...)
	if hasBestCheckpoint {
		keys = append(keys, "BEST CHECKPOINT")
		keys = append(keys, upper(metricsToDisplay)...)
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

		if displayHost {
			fmt.Fprintf(tw, "%s\t", exp.Host)
		}

		if displayUser {
			fmt.Fprintf(tw, "%s\t", exp.User)
		}

		// experiment params
		for _, heading := range paramsToDisplay {
			if val, ok := exp.Params[heading]; ok {
				fmt.Fprint(tw, val.ShortString(valueMaxLength, valueTruncate))
			}
			fmt.Fprintf(tw, "\t")
		}

		latestCheckpoint := ""
		if exp.LatestCheckpoint != nil {
			latestCheckpoint = fmt.Sprintf("%s (step %s)", exp.LatestCheckpoint.ShortID(), strconv.FormatInt(exp.LatestCheckpoint.Step, 10))
		}
		fmt.Fprintf(tw, "%s\t", latestCheckpoint)

		// latest checkpoint metrics
		for _, heading := range metricsToDisplay {
			val := ""
			if exp.LatestCheckpoint != nil {
				if v, ok := exp.LatestCheckpoint.Metrics[heading]; ok {
					val = v.ShortString(valueMaxLength, valueTruncate)
				}
			}
			fmt.Fprintf(tw, "%s\t", val)
		}

		bestCheckpoint := ""

		if exp.BestCheckpoint != nil {
			bestCheckpoint = fmt.Sprintf("%s (step %s)", exp.BestCheckpoint.ShortID(), strconv.FormatInt(exp.BestCheckpoint.Step, 10))
		}
		fmt.Fprintf(tw, "%s\t", bestCheckpoint)

		// best checkpoint metrics
		for _, heading := range metricsToDisplay {
			val := ""
			if exp.BestCheckpoint != nil {
				if v, ok := exp.BestCheckpoint.Metrics[heading]; ok {
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

// Get experiment params to display in list. If onlyChangedParams is true, only return
// params which have changed across experiments.
func getParamsToDisplay(experiments []*ListExperiment, all bool) []string {
	expHeadingSet := map[string]bool{}

	if all {
		for _, exp := range experiments {
			for key, val := range exp.Params {
				// Don't show objects in list view, because they're likely long and not very helpful
				if val.Type() == param.TypeObject {
					continue
				}
				expHeadingSet[key] = true
			}
		}
	} else {
		paramValues := param.ValueMap{}
		for _, exp := range experiments {
			for key, val := range exp.Params {
				// Don't show objects in list view, because they're likely long and not very helpful
				if val.Type() == param.TypeObject {
					continue
				}

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
	}

	return slices.StringKeys(expHeadingSet)
}

// Get metrics to display for each checkpoint shown in list
func getMetricsToDisplay(experiments []*ListExperiment, all bool) []string {
	metricsToDisplay := map[string]bool{}

	if all {
		for _, experiment := range experiments {
			checkpoint := experiment.BestCheckpoint
			if checkpoint == nil {
				checkpoint = experiment.LatestCheckpoint
			}
			if checkpoint != nil {
				for metric := range checkpoint.Metrics {
					metricsToDisplay[metric] = true
				}
			}
		}
	} else {
		for _, exp := range experiments {
			if exp.BestCheckpoint == nil {
				continue
			}
			metricsToDisplay[exp.BestCheckpoint.PrimaryMetric.Name] = true
		}
	}

	return slices.StringKeys(metricsToDisplay)
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
		running, err := proj.ExperimentIsRunning(exp.ID)
		if err != nil {
			return nil, err
		}
		listExperiment.LatestCheckpoint = exp.LatestCheckpoint()
		listExperiment.BestCheckpoint = exp.BestCheckpoint()
		listExperiment.NumCheckpoints = len(exp.Checkpoints)
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
