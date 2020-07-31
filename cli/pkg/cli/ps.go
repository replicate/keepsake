package cli

import (
	"os"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/list"
	"replicate.ai/cli/pkg/storage"
)

func newPsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ps",
		Short:   "List running experiments in this project",
		Aliases: []string{"processes"},
		RunE:    listRunningExperiments,
		Args:    cobra.NoArgs,
	}

	addStorageURLFlag(cmd)
	addListFormatFlag(cmd)

	return cmd
}

func listRunningExperiments(cmd *cobra.Command, args []string) error {
	storageURL, _, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	format, err := parseListFormatFlag(cmd)
	if err != nil {
		return err
	}
	store, err := storage.ForURL(storageURL)
	if err != nil {
		return err
	}
	return list.RunningExperiments(os.Stdout, store, format)
}
