package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/files"
	"replicate.ai/cli/pkg/interact"
	"replicate.ai/cli/pkg/storage"
)

func newCheckoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout <commit-id>",
		Short: "Check out code from a specific commit",
		RunE:  checkoutCommit,
		Args:  cobra.ExactArgs(1),
	}

	addStorageURLFlag(cmd)
	cmd.Flags().StringP("output-directory", "o", "", "Output directory (defaults to current working directory)")
	cmd.Flags().BoolP("force", "f", false, "Force checkout without prompt, even if the directory is not empty")

	return cmd
}

func checkoutCommit(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	outputDir, err := cmd.Flags().GetString("output-directory")
	if err != nil {
		return err
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	storageURL, _, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	store, err := storage.ForURL(storageURL)
	if err != nil {
		return err
	}

	exists, err := files.FileExists(outputDir)
	if err != nil {
		return err
	}
	if exists {
		isDir, err := files.IsDir(outputDir)
		if err != nil {
			return err
		}
		if !isDir {
			return fmt.Errorf("Checkout path %s is not a directory", outputDir)
		}
	} else {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("Failed to create directory %s, got error: %w", outputDir, err)
		}
	}

	isEmpty, err := files.DirIsEmpty(outputDir)
	if err != nil {
		return err
	}
	if !isEmpty && !force {
		// TODO(andreas): tell the user which files may get
		// overwritten, etc.
		doOverwrite, err := interact.InteractiveBool{
			Prompt:  fmt.Sprintf("The directory %s is not empty.\nThis checkout may overwrite existing files.\nDo you want to continue?", outputDir),
			Default: false,
		}.Read()
		if err != nil {
			return err
		}
		if !doOverwrite {
			console.Info("Aborting.")
			return nil
		}
	}

	proj := experiment.NewProject(store)
	com, err := proj.CommitFromPrefix(prefix)
	if err != nil {
		return err
	}

	// TODO(andreas): empty directory before getting new contents
	if err := store.GetDirectory(path.Join("commits", com.ID), outputDir); err != nil {
		return err
	}
	console.Info("Checked out %s to %s", com.ID, outputDir)
	return nil
}
