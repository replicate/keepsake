package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/cli/list"
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
	addListFormatFlags(cmd)

	return cmd
}

func listExperiments(cmd *cobra.Command, args []string) error {
	storageURL, _, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	format, allParams, err := parseListFormatFlags(cmd)
	if err != nil {
		return err
	}
	store, err := storage.ForURL(storageURL)
	if err != nil {
		return err
	}
	return list.Experiments(store, format, allParams)
}

func addListFormatFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("format", "f", "table", "Output format (table/json)")
	cmd.Flags().BoolP("all-params", "p", false, "Output all experiment params (by default, outputs only parameters that change between experiments)")
}

func parseListFormatFlags(cmd *cobra.Command) (format string, allParams bool, err error) {
	format, err = cmd.Flags().GetString("format")
	if err != nil {
		return "", false, err
	}
	validFormats := []string{list.FormatJSON, list.FormatTable}
	if !slices.ContainsString(validFormats, format) {
		return "", false, fmt.Errorf("%s is not a valid format. Valid formats are: %s", format, strings.Join(validFormats, ", "))
	}

	allParams, err = cmd.Flags().GetBool("all-params")
	if err != nil {
		return "", false, err
	}

	return format, allParams, nil
}
