package cli

import (
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/cli/list"
	"replicate.ai/cli/pkg/param"
)

func newPsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ps",
		Short:   "List running experiments in this project",
		Aliases: []string{"processes"},
		Run:     handleErrors(listRunningExperiments),
		Args:    cobra.NoArgs,
	}

	addStorageURLFlag(cmd)
	addListFormatFlags(cmd)
	addListFilterFlag(cmd)
	addListSortFlag(cmd)

	return cmd
}

func listRunningExperiments(cmd *cobra.Command, args []string) error {
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
	filters.SetExclusive("status", param.OperatorEqual, param.String("running"))
	store, err := getStorage(storageURL, projectDir)
	if err != nil {
		return err
	}
	return list.Experiments(store, format, allParams, filters, sortKey)
}
