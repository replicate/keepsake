package cli

import (
	"github.com/spf13/cobra"

	"github.com/replicate/replicate/go/pkg/cli/list"
	"github.com/replicate/replicate/go/pkg/param"
)

func newPsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ps",
		Short:   "List running experiments in this project",
		Aliases: []string{"processes"},
		Run:     handleErrors(listRunningExperiments),
		Args:    cobra.NoArgs,
	}

	addRepositoryURLFlag(cmd)
	addListFormatFlags(cmd)
	addListFilterFlag(cmd)
	addListSortFlag(cmd)

	return cmd
}

func listRunningExperiments(cmd *cobra.Command, args []string) error {
	repositoryURL, projectDir, err := getRepositoryURLFromFlagOrConfig(cmd)
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
	repo, err := getRepository(repositoryURL, projectDir)
	if err != nil {
		return err
	}
	return list.Experiments(repo, format, allParams, filters, sortKey)
}
