package list

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/xeonx/timeago"

	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/slices"
	"replicate.ai/cli/pkg/storage"
)

const FormatJSON = "json"
const FormatTable = "table"

func Experiments(store storage.Storage, format string) error {
	experiments, err := experiment.List(store)
	if err != nil {
		return err
	}
	// TODO: nil?
	// sort.Slice(experiments, func(i, j int) bool {
	// 	return experiments[i].LatestCommit().Created.Before(experiments[j].LatestCommit().Created)
	// })

	switch format {
	case FormatJSON:
		return outputJSON(experiments)
	case FormatTable:
		return outputTable(experiments)
	}
	return fmt.Errorf("Unknown format: %s", format)
}

func outputJSON(experiments []*experiment.Experiment) error {
	data, err := json.MarshalIndent(experiments, "", " ")
	if err != nil {
		return err
	}
	fmt.Println(string(data))
	return nil
}

func outputTable(experiments []*experiment.Experiment) error {
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
		if exp.IsRunning() {
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
		fmt.Fprintf(tw, "%d\t", len(exp.Commits))

		latestCommitTime := ""
		if exp.LatestCommit() != nil {
			latestCommitTime = formatTime(exp.LatestCommit().Created)
		}
		fmt.Fprintf(tw, "%s\t", latestCommitTime)
		for _, heading := range commitHeadings {
			if val, ok := exp.LatestCommit().Metrics[heading]; ok {
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

func getTableHeadings(experiments []*experiment.Experiment) (expHeadings []string, commitHeadings []string) {
	expHeadingSet := map[string]bool{}
	commitHeadingSet := map[string]bool{}

	for _, exp := range experiments {
		for key := range exp.Params {
			expHeadingSet[key] = true
		}
		if exp.LatestCommit() != nil {
			for key := range exp.LatestCommit().Metrics {
				commitHeadingSet[key] = true
			}
		}
	}

	return slices.StringKeys(expHeadingSet), slices.StringKeys(commitHeadingSet)
}
