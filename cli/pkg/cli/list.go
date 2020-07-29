package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/list"
	"replicate.ai/cli/pkg/slices"
	"replicate.ai/cli/pkg/storage"
)

func newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ls",
		Short:   "List experiments in this project",
		Aliases: []string{"list"},
		RunE:    listExperiments,
		Args:    cobra.NoArgs,
	}

	addStorageURLFlag(cmd)
	addListFormatFlag(cmd)

	return cmd
}

func listExperiments(cmd *cobra.Command, args []string) error {
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
	return list.Experiments(store, format)
}

func addListFormatFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("format", "f", "table", "Output format (table/json)")
}

func parseListFormatFlag(cmd *cobra.Command) (format string, err error) {
	format, err = cmd.Flags().GetString("format")
	if err != nil {
		return "", err
	}
	validFormats := []string{list.FormatJSON, list.FormatTable}
	if !slices.ContainsString(validFormats, format) {
		return "", fmt.Errorf("%s is not a valid format. Valid formats are: %s", format, strings.Join(validFormats, ", "))
	}

	return format, nil
}
