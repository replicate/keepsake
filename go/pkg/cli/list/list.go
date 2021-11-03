package list

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/replicate/keepsake/go/pkg/config"
	"github.com/replicate/keepsake/go/pkg/console"
	"github.com/replicate/keepsake/go/pkg/param"
	"github.com/replicate/keepsake/go/pkg/project"
	"github.com/replicate/keepsake/go/pkg/repository"
	"github.com/replicate/keepsake/go/pkg/slices"
)

type Format int

const (
	FormatJSON = iota
	FormatTable
	FormatQuiet
	FormatFullTable
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

// We should add some validation and better error messages, see https://github.com/replicate/keepsake/issues/340
func (exp *ListExperiment) GetValue(name string) param.Value {
	if name == "started" || name == "created" {
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
		return outputTable(listExperiments, all, true)
	case FormatFullTable:
		return outputTable(listExperiments, all, false)
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

func outputTable(experiments []*ListExperiment, all bool, truncate bool) error {
	if len(experiments) == 0 {
		console.Info("No experiments found")
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

	headings := []string{"EXPERIMENT", "STARTED", "STATUS"}
	if displayHost {
		headings = append(headings, "HOST")
	}
	if displayUser {
		headings = append(headings, "USER")
	}
	headings = append(headings, "PARAMS")
	if hasBestCheckpoint {
		headings = append(headings, "BEST CHECKPOINT")
	}
	headings = append(headings, "LATEST CHECKPOINT")

	for i, key := range headings {
		fmt.Fprintf(tw, "%s", key)
		if i < len(headings)-1 {
			fmt.Fprint(tw, "\t")
		}
	}
	fmt.Fprint(tw, "\n")

	for _, exp := range experiments {
		columns := []string{exp.ID[:7], console.FormatTime(exp.Created)}
		if exp.Running {
			columns = append(columns, "running")
		} else {
			columns = append(columns, "stopped")
		}

		if displayHost {
			columns = append(columns, exp.Host)
		}

		if displayUser {
			columns = append(columns, exp.User)
		}

		params := []string{}
		for _, key := range paramsToDisplay {
			if val, ok := exp.Params[key]; ok {
				if truncate {
					params = append(params, key+"="+val.ShortString(valueMaxLength, valueTruncate))
				} else {
					params = append(params, key+"="+val.String())
				}
			}
		}
		columns = append(columns, strings.Join(params, "\n"))

		if hasBestCheckpoint {
			bestCheckpoint := ""
			if exp.BestCheckpoint != nil {
				bestCheckpoint = displayCheckpoint(exp.BestCheckpoint, metricsToDisplay)
			}
			columns = append(columns, bestCheckpoint)
		}

		latestCheckpoint := ""
		if exp.LatestCheckpoint != nil {
			latestCheckpoint = displayCheckpoint(exp.LatestCheckpoint, metricsToDisplay)
		}
		columns = append(columns, latestCheckpoint)

		writeRow(tw, columns)

	}

	if err := tw.Flush(); err != nil {
		return err
	}

	return nil
}

func displayCheckpoint(checkpoint *project.Checkpoint, metricsToDisplay []string) string {
	out := []string{fmt.Sprintf("%s (step %s)", checkpoint.ShortID(), strconv.FormatInt(checkpoint.Step, 10))}

	for _, key := range metricsToDisplay {
		if v, ok := checkpoint.Metrics[key]; ok {
			out = append(out, key+"="+v.ShortString(valueMaxLength, valueTruncate))
		}
	}
	return strings.Join(out, "\n")
}

// Write columns with multiple rows to a tabwriter, adding correct padding tabs
// E.g. ["foo", "foo\nbar"] turns into:
// Fprint(w, "foo\tfoo")
// Fprint(w, "\tbar")
// Fprint(w, "\t")
func writeRow(w io.Writer, columns []string) {
	// Max number of lines in a column in this row
	numLines := 1
	for _, s := range columns {
		n := strings.Count(s, "\n") + 1
		if n > numLines {
			numLines = n
		}
	}
	// Add a blank line
	numLines += 1

	// Create sparse 2D array of lines/columns
	lines := make([][]string, numLines)
	for i := range lines {
		lines[i] = make([]string, len(columns))
	}

	// Put everything in its right spot
	for column := range columns {
		columnRows := strings.Split(columns[column], "\n")
		for row := range columnRows {
			lines[row][column] = columnRows[row]
		}
	}

	// Write the 2D array as lines and tabs
	for _, line := range lines {
		fmt.Fprintln(w, strings.Join(line, "\t"))
	}
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
