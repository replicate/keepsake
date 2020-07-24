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

func RunningExperiments(store storage.Storage, format string) error {
	commits, err := commit.ListCommits(store)
	if err != nil {
		return err
	}
	experiments := groupCommits(commits)

	running := []*GroupedExperiment{}
	for _, exp := range experiments {
		if exp.Running {
			running = append(running, exp)
		}
	}

	switch format {
	case FormatJSON:
		return outputJSON(running)
	case FormatTable:
		return outputTable(running)
	}
	return fmt.Errorf("Unknown format: %s", format)
}

func Experiments(store storage.Storage, format string) error {
	commits, err := commit.ListCommits(store)
	if err != nil {
		return err
	}
	experiments := groupCommits(commits)

	switch format {
	case FormatJSON:
		return outputJSON(experiments)
	case FormatTable:
		return outputTable(experiments)
	}
	return fmt.Errorf("Unknown format: %s", format)
}

func outputJSON(experiments []*GroupedExperiment) error {
	data, err := json.Marshal(experiments)
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputTable(experiments []*GroupedExperiment) error {
	if len(experiments) == 0 {
		fmt.Println("No experiments found")
		return nil
	}

	expHeadings, commitHeadings := getTableHeadings(experiments)

	tw := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

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

func groupCommits(commits []*commit.Commit) []*GroupedExperiment {
	expIDToCommits := make(map[string][]*commit.Commit)
	for _, com := range commits {
		expID := com.Experiment.ID
		if _, ok := expIDToCommits[expID]; !ok {
			expIDToCommits[expID] = []*commit.Commit{}
		}
		expIDToCommits[expID] = append(expIDToCommits[expID], com)
	}

	ret := []*GroupedExperiment{}
	for _, commits := range expIDToCommits {
		sort.Slice(commits, func(i, j int) bool {
			return commits[i].Created.Before(commits[j].Created)
		})
		latestCommit := commits[len(commits)-1]
		exp := latestCommit.Experiment
		groupedExperiment := GroupedExperiment{
			ID:           exp.ID,
			Params:       exp.Params,
			NumCommits:   len(commits),
			LatestCommit: latestCommit,
			Created:      exp.Created,
			Host:         exp.Host,
			User:         exp.User,
			Running:      exp.Running,
		}
		ret = append(ret, &groupedExperiment)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].LatestCommit.Created.Before(ret[j].LatestCommit.Created)
	})

	return ret
}
