package cli

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

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/slices"
)

var timezone = time.Local

type showOpts struct {
	all           bool
	json          bool
	repositoryURL string
}

func newShowCommand() *cobra.Command {
	var opts showOpts

	cmd := &cobra.Command{
		Use:   "show <experiment or checkpoint ID>",
		Short: "View information about an experiment or checkpoint",
		Run: handleErrors(func(cmd *cobra.Command, args []string) error {
			return show(opts, args, os.Stdout)
		}),
		Args: cobra.ExactArgs(1),
	}

	cmd.Flags().BoolVar(&opts.all, "all", false, "Show all information")
	cmd.Flags().BoolVar(&opts.json, "json", false, "Print output in JSON format")
	addRepositoryURLFlagVar(cmd, &opts.repositoryURL)

	return cmd
}

func show(opts showOpts, args []string, out io.Writer) error {
	prefix := args[0]
	repositoryURL, projectDir, err := getRepositoryURLFromStringOrConfig(opts.repositoryURL)
	if err != nil {
		return err
	}
	repo, err := getRepository(repositoryURL, projectDir)
	if err != nil {
		return err
	}
	proj := project.NewProject(repo, projectDir)
	result, err := proj.CheckpointOrExperimentFromPrefix(prefix)
	if err != nil {
		return err
	}

	au := getAurora()

	if opts.json {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if result.Checkpoint != nil {
			return enc.Encode(result.Checkpoint)
		}
		return enc.Encode(result.Experiment)
	}

	if result.Checkpoint != nil {
		return showCheckpoint(au, out, proj, result.Experiment, result.Checkpoint, opts.all)
	}
	return showExperiment(au, out, proj, result.Experiment, opts.all)
}

func showCheckpoint(au aurora.Aurora, out io.Writer, proj *project.Project, exp *project.Experiment, com *project.Checkpoint, all bool) error {
	experimentRunning, err := proj.ExperimentIsRunning(exp.ID)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "%s\n\n", au.Underline(au.Bold((fmt.Sprintf("Checkpoint: %s", com.ID)))))

	w := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	fmt.Fprintf(w, "Created:\t%s\n", com.Created.In(timezone).Format(time.RFC1123))
	fmt.Fprintf(w, "Path:\t%s\n", com.Path)
	fmt.Fprintf(w, "Step:\t%d\n", com.Step)

	fmt.Fprintf(w, "\t\n")
	fmt.Fprintf(w, "%s\t\n", au.Bold("Experiment"))

	fmt.Fprintf(w, "ID:\t%s\n", exp.ID)

	writeExperimentCommon(au, w, exp, experimentRunning, all)

	if err := writeCheckpointMetrics(au, w, proj, com); err != nil {
		return err
	}

	fmt.Fprintln(w)
	return w.Flush()
}

func showExperiment(au aurora.Aurora, out io.Writer, proj *project.Project, exp *project.Experiment, all bool) error {
	experimentRunning, err := proj.ExperimentIsRunning(exp.ID)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "%s\n\n", au.Underline(au.Bold(fmt.Sprintf("Experiment: %s", exp.ID))))

	w := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	writeExperimentCommon(au, w, exp, experimentRunning, all)
	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(out, "%s\n", au.Bold("Checkpoints"))

	bestCheckpoint := exp.BestCheckpoint()
	labelNames := []string{}

	cw := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	headings := []string{"ID", "STEP", "CREATED"}
	// FIXME(bfirsh): labels might change during experiment
	if len(exp.Checkpoints) != 0 {
		for label := range exp.Checkpoints[0].Metrics {
			labelNames = append(labelNames, label)
		}
		// TODO: put primary first
		sort.Strings(labelNames)
		for _, label := range labelNames {
			headings = append(headings, strings.ToUpper(label))
		}
	}
	fmt.Fprintf(cw, "%s\n", strings.Join(headings, "\t"))

	for _, checkpoint := range exp.Checkpoints {
		columns := []string{checkpoint.ShortID(), strconv.FormatInt(checkpoint.Step, 10), console.FormatTime(checkpoint.Created)}
		for _, label := range labelNames {
			val := checkpoint.Metrics[label]
			s := val.ShortString(10, 5)
			if bestCheckpoint != nil && bestCheckpoint.ID == checkpoint.ID && checkpoint.PrimaryMetric.Name == label {
				// TODO (bfirsh): this could be done more elegantly with some formatting
				s += " (best)"
			}
			columns = append(columns, s)
		}
		fmt.Fprintf(cw, "%s\n", strings.Join(columns, "\t"))
	}
	if err := cw.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(out, "\n")
	fmt.Fprintf(out, "To see more details about a checkpoint, run:\n")
	fmt.Fprintf(out, "  replicate show <checkpoint ID>\n")
	return nil
}

func isInterestingPythonPackage(pkg string) bool {
	switch pkg {
	case
		"cntk",
		"keras",
		"jax",
		"mxnet",
		"numpy",
		"pandas",
		"pytorch_lightning",
		"sklearn",
		"tensorflow",
		"torch":
		return true
	}
	return false
}

func writeExperimentCommon(au aurora.Aurora, w *tabwriter.Writer, exp *project.Experiment, experimentRunning bool, all bool) {
	fmt.Fprintf(w, "Created:\t%s\n", exp.Created.In(timezone).Format(time.RFC1123))
	if experimentRunning {
		fmt.Fprint(w, "Status:\trunning\n")
	} else {
		fmt.Fprint(w, "Status:\tstopped\n")
	}
	fmt.Fprintf(w, "Host:\t%s\n", exp.Host)
	fmt.Fprintf(w, "User:\t%s\n", exp.User)
	fmt.Fprintf(w, "Command:\t%s\n", exp.Command)

	fmt.Fprintf(w, "\t\n")
	fmt.Fprintf(w, "%s\t\n", au.Bold("Params"))

	if len(exp.Params) > 0 {
		for _, p := range exp.SortedParams() {
			fmt.Fprintf(w, "%s:\t%s\n", p.Name, p.Value.String())
		}
	} else {
		fmt.Fprintf(w, "%s\t\n", au.Faint("(none)"))
	}

	fmt.Fprintf(w, "\t\n")
	fmt.Fprintf(w, "%s\t\n", au.Bold("System"))
	fmt.Fprintf(w, "Python version:\t%s\n", exp.PythonVersion)

	fmt.Fprintf(w, "\t\n")
	fmt.Fprintf(w, "%s\t\n", au.Bold("Python packages"))
	if len(exp.PythonPackages) > 0 {
		packageNames := []string{}
		displayMore := false

		for name := range exp.PythonPackages {
			packageNames = append(packageNames, name)
		}
		sort.Strings(packageNames)

		if !all && len(packageNames) > 5 {
			interestingPackageNames := slices.FilterString(packageNames, isInterestingPythonPackage)
			if len(interestingPackageNames) == 0 {
				packageNames = packageNames[0:5]
			} else {
				packageNames = interestingPackageNames
			}
			displayMore = true
		}
		for _, name := range packageNames {
			fmt.Fprintf(w, "%s:\t%s\n", name, exp.PythonPackages[name])
		}
		if displayMore {
			moreCount := len(exp.PythonPackages) - len(packageNames)
			// This doesn't include a tab, breaking the alignment of subsequent tabwriter lines.
			// We should use a less shitty tabwriter that lets things span multiple columns.
			fmt.Fprintf(w, "%s\n", au.Faint(fmt.Sprintf("... and %d more. Use --all to view.", moreCount)))
		}
	} else {
		fmt.Fprintf(w, "%s\t\n", au.Faint("(none)"))
	}

	fmt.Fprintf(w, "\t\n")
}

func writeCheckpointMetrics(au aurora.Aurora, w *tabwriter.Writer, proj *project.Project, com *project.Checkpoint) error {
	fmt.Fprintf(w, "%s\t\n", au.Bold("Metrics"))
	metrics := com.SortedMetrics()
	if len(metrics) > 0 {
		for _, lab := range metrics {
			if com.PrimaryMetric != nil && com.PrimaryMetric.Name == lab.Name {
				fmt.Fprintf(w, "%s:\t%s (primary, %s)\n", lab.Name, lab.Value.String(), com.PrimaryMetric.Goal)
			} else {
				fmt.Fprintf(w, "%s:\t%s\n", lab.Name, lab.Value.String())
			}
		}
	} else {
		fmt.Fprintf(w, "%s\t\n", au.Faint("(none)"))
	}
	return nil
}
