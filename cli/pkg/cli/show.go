package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/storage"
)

var timezone = time.Local

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <experiment-or-commit-id>",
		Short: "View information about an experiment or commit",
		RunE:  show,
		Args:  cobra.ExactArgs(1),
	}

	// TODO(andreas): support json output
	addStorageURLFlag(cmd)

	return cmd
}

func show(cmd *cobra.Command, args []string) error {
	prefix := args[0]
	storageURL, _, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	store, err := storage.ForURL(storageURL)
	if err != nil {
		return err
	}
	proj := project.NewProject(store)
	result, err := proj.CommitOrExperimentFromPrefix(prefix)
	if err != nil {
		return err
	}

	au := getAurora()

	if result.Commit != nil {
		w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)
		return showCommit(au, w, proj, result.Commit)
	}
	return showExperiment(au, os.Stdout, proj, result.Experiment)
}

func showCommit(au aurora.Aurora, w *tabwriter.Writer, proj *project.Project, com *project.Commit) error {
	exp, err := proj.ExperimentByID(com.ExperimentID)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s\n\n", au.Yellow(fmt.Sprintf("Commit:\t%s", com.ID)))
	fmt.Fprintf(w, "Experiment:\t%s\n", exp.ID)

	writeExperimentCommon(au, w, exp)
	if err := writeCommitMetrics(au, w, proj, com); err != nil {
		return err
	}

	fmt.Fprintln(w)
	return w.Flush()
}

func showExperiment(au aurora.Aurora, out io.Writer, proj *project.Project, exp *project.Experiment) error {
	fmt.Fprintf(out, "%s\n\n", au.Underline(au.Bold(fmt.Sprintf("Experiment: %s", exp.ID))))

	w := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	writeExperimentCommon(au, w, exp)
	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(out, "%s\n", au.Bold("Commits"))

	commits, err := proj.ExperimentCommits(exp.ID)
	if err != nil {
		return err
	}
	bestCommit, err := proj.ExperimentBestCommit(exp.ID)
	if err != nil {
		return err
	}
	var primaryMetric *config.Metric
	if exp.Config != nil && exp.Config.PrimaryMetric() != nil {
		primaryMetric = exp.Config.PrimaryMetric()
	}
	labelNames := []string{}

	cw := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	headings := []string{"ID", "STEP", "CREATED"}
	// FIXME(bfirsh): labels might change during experiment
	if len(commits) != 0 {
		for label := range commits[0].Labels {
			labelNames = append(labelNames, label)
		}
		// TODO: put primary first
		sort.Strings(labelNames)
		for _, label := range labelNames {
			headings = append(headings, strings.ToUpper(label))
		}
	}
	fmt.Fprintf(cw, "%s\n", strings.Join(headings, "\t"))

	for _, commit := range commits {
		columns := []string{commit.ShortID(), strconv.Itoa(commit.Step), console.FormatTime(commit.Created)}
		for _, label := range labelNames {
			val := commit.Labels[label]
			s := val.ShortString(10, 5)
			if bestCommit != nil && bestCommit.ID == commit.ID && primaryMetric != nil && primaryMetric.Name == label {
				// TODO (bfirsh): this could be done more elegantly with some formatting
				s += " (best)"
			}
			columns = append(columns, s)
		}
		fmt.Fprintf(cw, "%s\n", strings.Join(columns, "\t"))
	}
	return cw.Flush()
}

func writeExperimentCommon(au aurora.Aurora, w *tabwriter.Writer, exp *project.Experiment) {
	fmt.Fprintf(w, "Created:\t%s\n", exp.Created.In(timezone).Format(time.RFC1123))
	fmt.Fprintf(w, "Host:\t%s\n", exp.Host)
	fmt.Fprintf(w, "User:\t%s\n", exp.User)

	fmt.Fprintf(w, "\t\n")
	fmt.Fprintf(w, "%s\t\n", au.Bold("Params"))

	if len(exp.Params) > 0 {
		for _, p := range exp.SortedParams() {
			fmt.Fprintf(w, "%s:\t%s\n", p.Name, p.Value.String())
		}
	} else {
		fmt.Fprintf(w, "%s\n", au.Faint("(none)"))
	}
	fmt.Fprintln(w)
}

func writeCommitMetrics(au aurora.Aurora, w *tabwriter.Writer, proj *project.Project, com *project.Commit) error {
	exp, err := proj.ExperimentByID(com.ExperimentID)
	if err != nil {
		return err
	}
	conf := exp.Config
	if conf == nil {
		conf = new(config.Config)
	}

	fmt.Fprintf(w, "%s\n", au.Bold("Metrics"))

	metricNameSet := map[string]bool{}
	if len(conf.Metrics) > 0 {
		for _, metric := range conf.Metrics {
			valueString := "(none)"
			value, ok := com.Labels[metric.Name]
			if ok {
				valueString = value.String()
			}
			primaryString := ""
			if metric.Primary {
				primaryString = "primary, "
			}
			fmt.Fprintf(w, "%s:\t%s (%sgoal: %s)\n", metric.Name, valueString, primaryString, metric.Goal)
			metricNameSet[metric.Name] = true
		}
	} else {
		fmt.Fprintf(w, "%s\n", au.Faint("(none)"))
	}
	labelNames := []string{}
	for name := range com.Labels {
		if _, ok := metricNameSet[name]; !ok {
			labelNames = append(labelNames, name)
		}
	}
	if len(labelNames) > 0 {
		fmt.Fprintf(w, "%s\n", au.Bold("Labels"))
		for _, lab := range com.SortedLabels() {
			if _, ok := metricNameSet[lab.Name]; !ok {
				fmt.Fprintf(w, "%s:\t%s\n", lab.Name, lab.Value.String())
			}
		}
	}
	return nil
}
