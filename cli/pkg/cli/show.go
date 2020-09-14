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

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/project"
)

var timezone = time.Local

func newShowCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <experiment or checkpoint ID>",
		Short: "View information about an experiment or checkpoint",
		RunE:  show,
		Args:  cobra.ExactArgs(1),
	}

	// TODO(andreas): support json output
	addStorageURLFlag(cmd)

	return cmd
}

func show(cmd *cobra.Command, args []string) error {
	prefix := args[0]
	storageURL, projectDir, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	store, err := getStorage(storageURL, projectDir)
	if err != nil {
		return err
	}
	proj := project.NewProject(store)
	result, err := proj.CheckpointOrExperimentFromPrefix(prefix)
	if err != nil {
		return err
	}

	au := getAurora()

	if result.Checkpoint != nil {
		return showCheckpoint(au, os.Stdout, proj, result.Checkpoint)
	}
	return showExperiment(au, os.Stdout, proj, result.Experiment)
}

func showCheckpoint(au aurora.Aurora, out io.Writer, proj *project.Project, com *project.Checkpoint) error {
	exp, err := proj.ExperimentByID(com.ExperimentID)
	if err != nil {
		return err
	}
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

	writeExperimentCommon(au, w, exp, experimentRunning)

	if err := writeCheckpointMetrics(au, w, proj, com); err != nil {
		return err
	}

	fmt.Fprintln(w)
	return w.Flush()
}

func showExperiment(au aurora.Aurora, out io.Writer, proj *project.Project, exp *project.Experiment) error {
	experimentRunning, err := proj.ExperimentIsRunning(exp.ID)
	if err != nil {
		return err
	}

	fmt.Fprintf(out, "%s\n\n", au.Underline(au.Bold(fmt.Sprintf("Experiment: %s", exp.ID))))

	w := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	writeExperimentCommon(au, w, exp, experimentRunning)
	if err := w.Flush(); err != nil {
		return err
	}

	fmt.Fprintf(out, "%s\n", au.Bold("Checkpoints"))

	checkpoints, err := proj.ExperimentCheckpoints(exp.ID)
	if err != nil {
		return err
	}
	bestCheckpoint, err := proj.ExperimentBestCheckpoint(exp.ID)
	if err != nil {
		return err
	}
	labelNames := []string{}

	cw := tabwriter.NewWriter(out, 0, 8, 2, ' ', 0)
	headings := []string{"ID", "STEP", "CREATED"}
	// FIXME(bfirsh): labels might change during experiment
	if len(checkpoints) != 0 {
		for label := range checkpoints[0].Metrics {
			labelNames = append(labelNames, label)
		}
		// TODO: put primary first
		sort.Strings(labelNames)
		for _, label := range labelNames {
			headings = append(headings, strings.ToUpper(label))
		}
	}
	fmt.Fprintf(cw, "%s\n", strings.Join(headings, "\t"))

	for _, checkpoint := range checkpoints {
		columns := []string{checkpoint.ShortID(), strconv.Itoa(checkpoint.Step), console.FormatTime(checkpoint.Created)}
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

func writeExperimentCommon(au aurora.Aurora, w *tabwriter.Writer, exp *project.Experiment, experimentRunning bool) {
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
}

func writeCheckpointMetrics(au aurora.Aurora, w *tabwriter.Writer, proj *project.Project, com *project.Checkpoint) error {
	fmt.Fprintf(w, "%s\t\n", au.Bold("Metrics"))
	metrics := com.SortedMetrics()
	if len(metrics) > 0 {
		for _, lab := range metrics {
			if com.PrimaryMetric.Name == lab.Name {
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
