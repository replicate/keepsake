package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"text/tabwriter"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/storage"
)

func newDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <id> <id>",
		Short: "Compare two experiments or commits",
		Long: `Compare two experiments or commits.

If an experiment ID is passed, it will pick the best commit from that experiment. If a primary metric is not defined in replicate.yaml, it will use the latest commit.`,
		RunE: diffCommits,
		Args: cobra.ExactArgs(2),
	}

	// TODO(andreas): support json output
	addStorageURLFlag(cmd)

	return cmd
}

func diffCommits(cmd *cobra.Command, args []string) error {
	// TODO(andreas): generalize to >2 commits/checkpoints

	prefix1 := args[0]
	prefix2 := args[1]

	storageURL, _, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	store, err := storage.ForURL(storageURL)
	if err != nil {
		return err
	}
	proj := project.NewProject(store)
	au := getAurora()
	return printDiff(os.Stdout, au, proj, prefix1, prefix2)
}

func printDiff(out io.Writer, au aurora.Aurora, proj *project.Project, prefix1 string, prefix2 string) error {
	com1, err := loadCommit(proj, prefix1)
	if err != nil {
		return err
	}
	com2, err := loadCommit(proj, prefix2)
	if err != nil {
		return err
	}
	exp1, err := proj.ExperimentByID(com1.ExperimentID)
	if err != nil {
		return err
	}
	exp2, err := proj.ExperimentByID(com2.ExperimentID)
	if err != nil {
		return err
	}

	// min width for 3 columns in 78 char terminal
	w := tabwriter.NewWriter(out, 78/3, 8, 2, ' ', 0)

	fmt.Fprintf(w, "Checkpoint:\t%s\t%s\n", com1.ID, com2.ID)
	fmt.Fprintf(w, "Experiment:\t%s\t%s\n", com1.ExperimentID, com2.ExperimentID)
	w.Flush()

	fmt.Fprintf(w, "\n")

	fmt.Fprintf(w, "%s\n", au.Bold("Params"))
	printMapDiff(w, au, paramMapToStringMap(exp1.Params), paramMapToStringMap(exp2.Params))
	fmt.Fprintln(w)

	metrics1 := map[string]*param.Value{}
	metrics2 := map[string]*param.Value{}
	if exp1.HasMetrics() || exp2.HasMetrics() {
		fmt.Fprintf(w, "%s\n", au.Bold("Metrics"))
		for _, metric := range exp1.Config.Metrics {
			if value, ok := com1.Labels[metric.Name]; ok {
				metrics1[metric.Name] = value
			}
		}
		for _, metric := range exp2.Config.Metrics {
			if value, ok := com2.Labels[metric.Name]; ok {
				metrics2[metric.Name] = value
			}
		}
		printMapDiff(w, au, paramMapToStringMap(metrics1), paramMapToStringMap(metrics2))
		fmt.Fprintln(w)
	}

	fmt.Fprintf(w, "%s\n", au.Bold("Labels"))
	labels1 := map[string]*param.Value{}
	labels2 := map[string]*param.Value{}
	for name, label := range com1.Labels {
		if _, ok := metrics1[name]; !ok {
			labels1[name] = label
		}
	}
	for name, label := range com2.Labels {
		if _, ok := metrics2[name]; !ok {
			labels2[name] = label
		}
	}
	printMapDiff(w, au, paramMapToStringMap(labels1), paramMapToStringMap(labels2))
	fmt.Fprintln(w)
	return w.Flush()
}

func printMapDiff(w *tabwriter.Writer, au aurora.Aurora, map1, map2 map[string]string) {
	diffMap := mapString(map1, map2)

	// sort the keys
	type keyVal struct {
		key   string
		value []*string
	}
	keyVals := []keyVal{}
	for k, v := range diffMap {
		keyVals = append(keyVals, keyVal{k, v})
	}
	sort.Slice(keyVals, func(i, j int) bool {
		return keyVals[i].key < keyVals[j].key
	})

	if len(keyVals) > 0 {
		for _, kv := range keyVals {
			left := au.Faint("(not set)").String()
			right := au.Faint("(not set)").String()
			if kv.value[0] != nil {
				left = *(kv.value[0])
			}
			if kv.value[1] != nil {
				right = *(kv.value[1])
			}
			fmt.Fprintf(w, "%s:\t%s\t%s\n", kv.key, left, right)
		}
	} else {
		fmt.Fprintf(w, "%s\n", au.Faint("(no difference)"))
	}
}

func paramMapToStringMap(params map[string]*param.Value) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		result[k] = v.String()
	}
	return result
}

// loadCommit returns a commit given a prefix. If the prefix matches a
// commit, that is returned. If the prefix matches an experiment, it
// returns the best commit if a primary metric is defined in config,
// otherwise the latest commit.
func loadCommit(proj *project.Project, prefix string) (*project.Commit, error) {
	obj, err := proj.CommitOrExperimentFromPrefix(prefix)
	if err != nil {
		return nil, err
	}
	if obj.Commit != nil {
		return obj.Commit, nil
	}
	exp := obj.Experiment
	if exp.Config != nil && exp.Config.PrimaryMetric() != nil {
		return proj.ExperimentBestCommit(exp.ID)
	}
	return proj.ExperimentLatestCommit(exp.ID)
}

// mapString takes two maps of strings and returns a single map with two values
// where the values are different. If only one map has a key, then the map
// without the value will be marked as nil
//
// e.g.
// >>> mapString({"layers": "2", "foo": "bar"}, {"layers": "4"})
// {
//    "foo": ["bar", nil],
//	  "layers": ["2", "4"]
// }
//
func mapString(left, right map[string]string) map[string][]*string {
	result := make(map[string][]*string)
	for k, v := range left {
		if _, ok := right[k]; ok {
			// Key in both left and right
			if v != right[k] {
				// copy so pointers are unique
				v2 := v
				s := right[k]
				result[k] = []*string{&v2, &s}
			}
		} else {
			// Key just in left
			v2 := v
			result[k] = []*string{&v2, nil}
		}
	}
	for k, v := range right {
		// Key just in right
		if _, ok := left[k]; !ok {
			v2 := v
			result[k] = []*string{nil, &v2}
		}
	}
	return result
}
