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
	ID             string                  `json:"id"`
	StartTimestamp float64                 `json:"start_timestamp"`
	Params         map[string]*param.Value `json:"params"`
	NumCommits     int                     `json:"num_commits"`
	LatestCommit   *commit.Commit          `json:"latest_commit"`
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

	keys := []string{"experiment", "started"}
	for _, heading := range expHeadings {
		keys = append(keys, heading)
	}
	keys = append(keys, "commits", "latest")
	for _, heading := range commitHeadings {
		keys = append(keys, heading)
	}
	for i, key := range keys {
		fmt.Fprintf(tw, "%s", key)
		if i < len(keys)-1 {
			fmt.Fprint(tw, "\t")
		}
	}
	fmt.Fprint(tw, "\n")

	for _, exp := range experiments {
		fmt.Fprintf(tw, "%s\t", exp.ID[:7])
		fmt.Fprintf(tw, "%s\t", formatTimestamp(exp.StartTimestamp))
		for _, heading := range expHeadings {
			if val, ok := exp.Params[heading]; ok {
				fmt.Fprintf(tw, "%v\t", val)
			}
		}
		fmt.Fprintf(tw, "%d\t", exp.NumCommits)
		fmt.Fprintf(tw, "%s\t", formatTimestamp(exp.LatestCommit.Timestamp))
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

func formatTimestamp(timestamp float64) string {
	return timeago.English.Format(time.Unix(int64(timestamp), 0))
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
			return commits[i].Timestamp < commits[j].Timestamp
		})
		latestCommit := commits[len(commits)-1]
		exp := latestCommit.Experiment
		groupedExperiment := GroupedExperiment{
			ID:             exp.ID,
			Params:         exp.Params,
			NumCommits:     len(commits),
			LatestCommit:   latestCommit,
			StartTimestamp: exp.Timestamp,
		}
		ret = append(ret, &groupedExperiment)
	}

	sort.Slice(ret, func(i, j int) bool {
		return ret[i].LatestCommit.Timestamp < ret[j].LatestCommit.Timestamp
	})

	return ret
}
