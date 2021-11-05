package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replicate/keepsake/go/pkg/cli/list"
	"github.com/replicate/keepsake/go/pkg/param"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List experiments in this project",
		Aliases: []string{"list"},
		Run:     handleErrors(listExperiments),
		Args:    cobra.NoArgs,
		Example: `List all experiments in the current project:
$ keepsake ls

List experiments that have run for 50 steps or less:
$ keepsake ls --filter "step <= 50"

List experiments where the parameter "optimizer" is "adam" and
the best "accuracy" metric is greater than 0.8:
$ keepsake ls --filter "optimizer = adam" --filter "accuracy > 0.8"

Sort all stopped experiments by the metric "val_loss":
$ keepsake ls --sort "val_loss" --filter "status = stopped"
`,
	}

	addRepositoryURLFlag(cmd)
	addListFormatFlags(cmd)
	addListFilterFlag(cmd)
	addListSortFlag(cmd)

	return cmd
}

func listExperiments(cmd *cobra.Command, args []string) error {
	repositoryURL, projectDir, err := getRepositoryURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	format, all, err := parseListFormatFlags(cmd)
	if err != nil {
		return err
	}
	filters, err := parseListFilterFlag(cmd)
	if err != nil {
		return err
	}
	sortKey, err := parseListSortFlag(cmd)
	if err != nil {
		return err
	}
	repo, err := getRepository(repositoryURL, projectDir)
	if err != nil {
		return err
	}
	return list.Experiments(repo, format, all, filters, sortKey)
}

func addListFormatFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Print output in JSON format")
	cmd.Flags().Bool("all", false, "Output all params and metrics. Default: only params/metrics that differ")
	cmd.Flags().Bool("full", false, "Do not truncate parameter values.")
	cmd.Flags().BoolP("quiet", "q", false, "Only print experiment IDs")
}

// FIXME(bfirsh): use an opts struct and the "Var" version of flag functions to get rid of this
func parseListFormatFlags(cmd *cobra.Command) (format list.Format, all bool, err error) {
	json, err := cmd.Flags().GetBool("json")
	if err != nil {
		return 0, false, err
	}
	if json {
		format = list.FormatJSON
	} else {
		format = list.FormatTable
	}

	full, err := cmd.Flags().GetBool("full")
	if err != nil {
		return 0, false, err
	}
	if full && format == list.FormatJSON {
		return 0, false, fmt.Errorf("Cannot use the --full flag in combination with --json")
	}
	if full {
		format = list.FormatFullTable
	}

	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return 0, false, err
	}
	if quiet && format == list.FormatJSON {
		return 0, false, fmt.Errorf("Cannot use the --quiet flag in combination with --json")
	}
	if quiet && full {
		return 0, false, fmt.Errorf("Cannot use the --quiet flag in combination with --full")
	}

	all, err = cmd.Flags().GetBool("all")
	if err != nil {
		return 0, false, err
	}
	if quiet && all {
		return 0, false, fmt.Errorf("Cannot use the --quiet flag in combination with --all")
	}
	if quiet {
		format = list.FormatQuiet
	}

	return format, all, nil
}

func addListFilterFlag(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("filter", "f", []string{}, "Filters (format: \"<name> <operator> <value>\")")
}

// The filter names ought to be validated, see https://github.com/replicate/keepsake/issues/340
func parseListFilterFlag(cmd *cobra.Command) (*param.Filters, error) {
	filtersStr, err := cmd.Flags().GetStringArray("filter")
	if err != nil {
		return nil, err
	}
	if len(filtersStr) > 0 {
		filters, err := param.MakeFilters(filtersStr)
		if err != nil {
			return nil, err
		}
		return filters, nil
	}
	return new(param.Filters), nil
}

func addListSortFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("sort", "s", "created", "Sort key. Suffix with '-desc' for descending sort, e.g. --sort=created-desc")
}

func parseListSortFlag(cmd *cobra.Command) (*param.Sorter, error) {
	sortString, err := cmd.Flags().GetString("sort")
	if err != nil {
		return nil, err
	}
	return param.NewSorter(sortString), nil
}
