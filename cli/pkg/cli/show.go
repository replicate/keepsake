package cli

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/storage"
)

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <experiment-or-commit-id>",
		Short: "View information about an experiment or commit",
		RunE:  show,
		Args:  cobra.ExactArgs(1),
	}

	// TODO(andreas): support json output
	// TOOD(andreas): --all-commits
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
	w := tabwriter.NewWriter(os.Stdout, 0, 8, 2, ' ', 0)

	if result.Commit != nil {
		return showCommit(au, w, proj, result.Commit)
	}
	return showExperiment(au, w, proj, result.Experiment)
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

func showExperiment(au aurora.Aurora, w *tabwriter.Writer, proj *project.Project, exp *project.Experiment) error {
	fmt.Fprintf(w, "%s\n\n", au.Yellow(fmt.Sprintf("Experiment:\t%s", exp.ID)))

	writeExperimentCommon(au, w, exp)

	latest, err := proj.ExperimentLatestCommit(exp.ID)
	if err != nil {
		return err
	}
	fmt.Fprintf(w, "%s\n", au.Bold(fmt.Sprintf("Latest commit\t%s", latest.ID)))
	if err := writeCommitMetrics(au, w, proj, latest); err != nil {
		return err
	}

	best, err := proj.ExperimentBestCommit(exp.ID)
	if err != nil {
		return err
	}
	if best != nil {
		fmt.Fprintf(w, "%s\n", au.Bold(fmt.Sprintf("Best commit\t%s", best.ID)))
		if err := writeCommitMetrics(au, w, proj, best); err != nil {
			return err
		}
	}

	fmt.Fprintln(w)
	return w.Flush()
}

func writeExperimentCommon(au aurora.Aurora, w *tabwriter.Writer, exp *project.Experiment) {
	fmt.Fprintf(w, "%s\n", au.Bold("Params"))
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
