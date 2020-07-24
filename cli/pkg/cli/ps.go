package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/global"
	"replicate.ai/cli/pkg/list"
	"replicate.ai/cli/pkg/slices"
	"replicate.ai/cli/pkg/storage"
)

func newPsCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ps",
		Short:   "List running experiments in this project",
		Aliases: []string{"processes"},
		RunE:    listRunningExperiments,
		Args:    cobra.ExactArgs(0),
	}

	cmd.Flags().StringP("format", "f", "table", "Output format (table/json)")

	return cmd
}

func listRunningExperiments(cmd *cobra.Command, args []string) error {
	var storageURL string
	// FIXME (bfirsh): perhaps better as a flag? (just putting this here to remind ourselves to have this discussion)
	if len(args) == 1 {
		storageURL = args[0]
	} else {
		conf, _, err := config.FindConfigInWorkingDir(global.SourceDirectory)
		if err != nil {
			return err
		}
		storageURL = conf.Storage
	}

	format, err := cmd.Flags().GetString("format")
	if err != nil {
		return err
	}
	validFormats := []string{list.FormatJSON, list.FormatTable}
	if !slices.ContainsString(validFormats, format) {
		return fmt.Errorf("%s is not a valid format. Valid formats are: %s", format, strings.Join(validFormats, ", "))
	}

	store, err := storage.ForURL(storageURL)
	if err != nil {
		return err
	}

	return list.RunningExperiments(store, format)
}
