package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/cli/list"
	"replicate.ai/cli/pkg/param"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List experiments in this project",
		Aliases: []string{"list"},
		Run:     handleErrors(listExperiments),
		Args:    cobra.NoArgs,
	}

	addStorageURLFlag(cmd)
	addListFormatFlags(cmd)
	addListFilterFlag(cmd)
	addListSortFlag(cmd)

	return cmd
}

func listExperiments(cmd *cobra.Command, args []string) error {
	storageURL, projectDir, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	format, allParams, err := parseListFormatFlags(cmd)
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
	store, err := getStorage(storageURL, projectDir)
	if err != nil {
		return err
	}
	return list.Experiments(store, format, allParams, filters, sortKey)
}

func addListFormatFlags(cmd *cobra.Command) {
	cmd.Flags().Bool("json", false, "Print output in JSON format")
	cmd.Flags().BoolP("all-params", "p", false, "Output all experiment params (by default, outputs only parameters that change between experiments)")
	cmd.Flags().BoolP("quiet", "q", false, "Only print experiment IDs")
}

// FIXME(bfirsh): use an opts struct and the "Var" version of flag functions to get rid of this
func parseListFormatFlags(cmd *cobra.Command) (format list.Format, allParams bool, err error) {
	json, err := cmd.Flags().GetBool("json")
	if err != nil {
		return 0, false, err
	}
	if json {
		format = list.FormatJSON
	} else {
		format = list.FormatTable
	}

	quiet, err := cmd.Flags().GetBool("quiet")
	if err != nil {
		return 0, false, err
	}
	if quiet && format == list.FormatJSON {
		return 0, false, fmt.Errorf("Cannot use the --quiet flag in combination with --json")
	}

	allParams, err = cmd.Flags().GetBool("all-params")
	if err != nil {
		return 0, false, err
	}
	if quiet && allParams {
		return 0, false, fmt.Errorf("Cannot use the --quiet flag in combination with --all-params")
	}
	if quiet {
		format = list.FormatQuiet
	}

	return format, allParams, nil
}

func addListFilterFlag(cmd *cobra.Command) {
	cmd.Flags().StringArrayP("filter", "f", []string{}, "Filters (format: \"<name> <operator> <value>\")")
}

// TODO(andreas): validate filter name
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
	cmd.Flags().StringP("sort", "s", "started", "Sort key. Suffix with '-desc' for descending sort, e.g. --sort=started-desc")
}

func parseListSortFlag(cmd *cobra.Command) (*param.Sorter, error) {
	sortString, err := cmd.Flags().GetString("sort")
	if err != nil {
		return nil, err
	}
	return param.NewSorter(sortString), nil
}
