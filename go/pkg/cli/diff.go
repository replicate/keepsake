package cli

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/param"
	"github.com/replicate/replicate/go/pkg/project"
)

func newDiffCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff <ID> <ID>",
		Short: "Compare two experiments or checkpoints",
		Long: `Compare two experiments or checkpoints.

If an experiment ID is passed, it will pick the best checkpoint from that experiment. If a primary metric is not defined in replicate.yaml, it will use the latest checkpoint.`,
		Run:  handleErrors(diffCheckpoints),
		Args: cobra.ExactArgs(2),
	}

	// TODO(andreas): support json output
	addStorageURLFlag(cmd)

	return cmd
}

func diffCheckpoints(cmd *cobra.Command, args []string) error {
	// TODO(andreas): generalize to >2 checkpoints/experiments
	// TODO(bfirsh): it probably makes sense to refactor this to diff param.Values instead of strings at some point.
	// that way we can do interesting stuff like diff JSON structures, using param.Value comparison methods, ShortString, etc.

	prefix1 := args[0]
	prefix2 := args[1]

	storageURL, projectDir, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	store, err := getStorage(storageURL, projectDir)
	if err != nil {
		return err
	}
	proj := project.NewProject(store)
	au := getAurora()
	return printDiff(os.Stdout, au, proj, prefix1, prefix2)
}

// TODO: implement this as a thing in console
func br(w *tabwriter.Writer) {
	fmt.Fprintf(w, "\t\t\n")
}

func heading(w *tabwriter.Writer, au aurora.Aurora, text string) {
	fmt.Fprintf(w, "%s\t\t\n", au.Bold(text))
}

func printDiff(out io.Writer, au aurora.Aurora, proj *project.Project, prefix1 string, prefix2 string) error {
	exp1, com1, err := loadCheckpoint(proj, prefix1)
	if err != nil {
		return err
	}
	exp2, com2, err := loadCheckpoint(proj, prefix2)
	if err != nil {
		return err
	}

	// min width for 3 columns in 78 char terminal
	w := tabwriter.NewWriter(out, 78/3, 8, 2, ' ', 0)

	heading(w, au, "Experiment")
	fmt.Fprintf(w, "ID:\t%s\t%s\n", exp1.ShortID(), exp2.ShortID())
	// HACK: don't show "no differences" if it's the same experiment, but still show ID because that's useful
	if exp1.ID != exp2.ID {
		printMapDiff(w, au, experimentToMap(exp1), experimentToMap(exp2))
	}

	br(w)

	heading(w, au, "Params")
	printMapDiff(w, au, paramMapToStringMap(exp1.Params), paramMapToStringMap(exp2.Params))
	br(w)

	heading(w, au, "Python Packages")
	printMapDiff(w, au, exp1.PythonPackages, exp2.PythonPackages)
	br(w)

	heading(w, au, "Checkpoint")
	fmt.Fprintf(w, "ID:\t%s\t%s\n", com1.ShortID(), com2.ShortID())
	printMapDiff(w, au, checkpointToMap(com1), checkpointToMap(com2))
	br(w)

	heading(w, au, "Metrics")
	// TODO(bfirsh): put primary metric first
	printMapDiff(w, au, paramMapToStringMap(com1.Metrics), paramMapToStringMap(com2.Metrics))
	br(w)

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
			left := "(not set)"
			right := "(not set)"
			if kv.value[0] != nil {
				left = *(kv.value[0])
			}
			if kv.value[1] != nil {
				right = *(kv.value[1])
			}
			// Truncate to 50, which seems ball-park sensible figure to make this fit in a wide terminal
			// At some point when we have a clever responsive tabwriter, we can adjust this based on terminal width!
			fmt.Fprintf(w, "%s:\t%s\t%s\n", kv.key, param.Truncate(left, 50), param.Truncate(right, 50))
		}
	} else {
		fmt.Fprintf(w, "%s\t\t\n", au.Faint("(no difference)"))
	}
}

// Returns a map of checkpoint things we want to show in diff
func checkpointToMap(checkpoint *project.Checkpoint) map[string]string {
	return map[string]string{
		"Step":    strconv.Itoa(checkpoint.Step),
		"Created": checkpoint.Created.In(timezone).Format(time.RFC1123),
		"Path":    checkpoint.Path,
	}
}

// Returns a map of checkpoint things we want to show in diff
func experimentToMap(exp *project.Experiment) map[string]string {
	return map[string]string{
		"Created": exp.Created.In(timezone).Format(time.RFC1123),
		"Host":    exp.Host,
		"User":    exp.User,
		"Command": exp.Command,
	}
}

func paramMapToStringMap(params param.ValueMap) map[string]string {
	result := make(map[string]string)
	for k, v := range params {
		result[k] = v.String()
	}
	return result
}

// loadCheckpoint returns a checkpoint given a prefix. If the prefix matches a
// checkpoint, that is returned. If the prefix matches an experiment, it
// returns the best checkpoint if a primary metric is defined in config,
// otherwise the latest checkpoint.
func loadCheckpoint(proj *project.Project, prefix string) (*project.Experiment, *project.Checkpoint, error) {
	obj, err := proj.CheckpointOrExperimentFromPrefix(prefix)
	if err != nil {
		return nil, nil, err
	}
	if obj.Checkpoint != nil {
		return obj.Experiment, obj.Checkpoint, nil
	}
	exp := obj.Experiment

	// First, try getting best checkpoint
	checkpoint := exp.BestCheckpoint()
	if checkpoint != nil {
		console.Info("%q matches an experiment, picking the best checkpoint", prefix)
		return exp, checkpoint, nil
	}

	// If there is no best checkpoint and no error, then no primary metric has been set,
	// so fall back to picking latest checkpoint
	console.Info("%q is an experiment, picking the latest checkpoint", prefix)
	checkpoint = exp.LatestCheckpoint()
	if checkpoint == nil {
		return nil, nil, fmt.Errorf("Could not pick best checkpoint for experiment %q: it does not have any checkpoints.", exp.ShortID())
	}
	return exp, checkpoint, nil
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
