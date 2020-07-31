package list

import (
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/xeonx/timeago"

	"replicate.ai/cli/pkg/commit"
	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/slices"
	"replicate.ai/cli/pkg/storage"
)

const FormatJSON = "json"
const FormatTable = "table"

type GroupedExperiment struct {
	ID           string                  `json:"id"`
	Created      time.Time               `json:"created"`
	Params       map[string]*param.Value `json:"params"`
	NumCommits   int                     `json:"num_commits"`
	LatestCommit *commit.Commit          `json:"latest_commit"`
	User         string                  `json:"user"`
	Host         string                  `json:"host"`
	Running      bool                    `json:"running"`
}

func RunningExperiments(out io.Writer, store storage.Storage, format string) error {
	experiments, err := experiment.List(store)
	if err != nil {
		return err
	}

	commits, err := commit.ListCommits(store)
	if err != nil {
		return err
	}

	groupedExperiments := groupExperiments(store, experiments, commits)

	running := []*GroupedExperiment{}
	for _, exp := range groupedExperiments {
		if exp.Running {
			running = append(running, exp)
		}
	}

	switch format {
	case FormatJSON:
		return outputJSON(out, running)
	case FormatTable:
		return outputTable(out, running)
	}
	return fmt.Errorf("Unknown format: %s", format)
}

func Experiments(out io.Writer, store storage.Storage, format string) error {
	experiments, err := experiment.List(store)
	if err != nil {
		return err
	}

	commits, err := commit.ListCommits(store)
	if err != nil {
		return err
	}
	groupedExperiments := groupExperiments(store, experiments, commits)

	switch format {
	case FormatJSON:
		return outputJSON(out, groupedExperiments)
	case FormatTable:
		return outputTable(out, groupedExperiments)
	}
	return fmt.Errorf("Unknown format: %s", format)
}

func outputJSON(out io.Writer, experiments []*GroupedExperiment) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	return enc.Encode(experiments)
}

func outputTable(out io.Writer, experiments []*GroupedExperiment) error {
	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	expHeadings, commitHeadings := getTableHeadings(experiments)

	tw := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)

	keys := []string{"experiment", "started", "status", "host", "user"}
	keys = append(keys, expHeadings...)
	keys = append(keys, "commits", "latest")
	keys = append(keys, commitHeadings...)
	for i, key := range keys {
		fmt.Fprintf(tw, "%s", key)
		if i < len(keys)-1 {
			fmt.Fprint(tw, "\t")
		}
	}
	fmt.Fprint(tw, "\n")

	for _, exp := range experiments {
		fmt.Fprintf(tw, "%s\t", exp.ID[:7])
		fmt.Fprintf(tw, "%s\t", formatTime(exp.Created))
		if exp.Running {
			fmt.Fprint(tw, "running\t")
		} else {
			fmt.Fprint(tw, "stopped\t")
		}
		fmt.Fprintf(tw, "%s\t", exp.Host)
		fmt.Fprintf(tw, "%s\t", exp.User)
		for _, heading := range expHeadings {
			if val, ok := exp.Params[heading]; ok {
				fmt.Fprintf(tw, "%v\t", val)
			}
		}
		fmt.Fprintf(tw, "%d\t", exp.NumCommits)
		fmt.Fprintf(tw, "%s\t", formatTime(exp.LatestCommit.Created))
		for _, heading := range commitHeadings {
			if val, ok := exp.LatestCommit.Metrics[heading]; ok {
				fmt.Fprintf(tw, "%v\t", val)
			}
		}
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

func getTableHeadings(experiments []*GroupedExperiment) (expHeadings []string, commitHeadings []string) {
	expHeadingSet := map[string]bool{}
	commitHeadingSet := map[string]bool{}

	for _, exp := range experiments {
		for key := range exp.Params {
			expHeadingSet[key] = true
		}
		for key := range exp.LatestCommit.Metrics {
			commitHeadingSet[key] = true
		}
	}

	return slices.StringKeys(expHeadingSet), slices.StringKeys(commitHeadingSet)
}

func groupExperiments(store storage.Storage, experiments []*experiment.Experiment, commits []*commit.Commit) []*GroupedExperiment {
	expIDToCommits := make(map[string][]*commit.Commit)
	for _, com := range commits {
		expID := com.ExperimentID
		if _, ok := expIDToCommits[expID]; !ok {
			expIDToCommits[expID] = []*commit.Commit{}
		}
		expIDToCommits[expID] = append(expIDToCommits[expID], com)
	}

	ret := []*GroupedExperiment{}

	for _, exp := range experiments {
		groupedExperiment := GroupedExperiment{
			ID:      exp.ID,
			Params:  exp.Params,
			Created: exp.Created,
			Host:    exp.Host,
			User:    exp.User,
		}

		commits, ok := expIDToCommits[exp.ID]
		if ok {
			sort.Slice(commits, func(i, j int) bool {
				return commits[i].Created.Before(commits[j].Created)
			})
			groupedExperiment.LatestCommit = commits[len(commits)-1]
			groupedExperiment.NumCommits = len(commits)
		}

		heartbeat, err := experiment.LoadHeartbeat(store, exp.ID)
		// TODO: handle errors other than heartbeat not existing
		if err == nil {
			groupedExperiment.Running = heartbeat.IsRunning()
		}

		ret = append(ret, &groupedExperiment)

	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].LatestCommit.Created.Before(ret[j].LatestCommit.Created)
	})

	return ret

}
