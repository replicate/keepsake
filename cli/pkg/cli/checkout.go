package cli

import (
	"fmt"
	"os"
	"path"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/files"
	"replicate.ai/cli/pkg/interact"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/storage"
)

func newCheckoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout <commit-id>",
		Short: "Copy files from a commit into the project directory",
		RunE:  checkoutCommit,
		Args:  cobra.ExactArgs(1),
	}

	addStorageURLFlag(cmd)
	cmd.Flags().StringP("output-directory", "o", "", "Output directory (defaults to working directory or directory with replicate.yaml in it)")
	cmd.Flags().BoolP("force", "f", false, "Force checkout without prompt, even if the directory is not empty")

	return cmd
}

func checkoutCommit(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	outputDir, err := cmd.Flags().GetString("output-directory")
	if err != nil {
		return err
	}
	// TODO(andreas): add test for case where --output-directory is omitted
	if outputDir == "" {
		outputDir, err = getSourceDir()
		if err != nil {
			return err
		}
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
			return fmt.Errorf("Checkout path %q is not a directory", outputDir)
		}
	} else {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("Failed to create directory %q, got error: %w", outputDir, err)
		}
	}

	proj := project.NewProject(store)
	com, err := proj.CommitFromPrefix(prefix)
	if err != nil {
		return err
	}

	isEmpty, err := files.DirIsEmpty(outputDir)
	if err != nil {
		return err
	}
	if !isEmpty && !force {
		console.Warn("The directory %q is not empty.", outputDir)
		console.Warn("%s Make sure they're saved in Git or Replicate so they're safe!", aurora.Bold("This checkout may overwrite existing files."))
		fmt.Println()
		// TODO(andreas): tell the user which files may get
		// overwritten, etc.
		doOverwrite, err := interact.InteractiveBool{
			Prompt:  "Do you want to continue?",
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

	// TODO(andreas): empty directory before getting new contents
	if err := store.GetDirectory(path.Join("experiments", com.ExperimentID), outputDir); err != nil {
		return err
	}
	// Overlay commit on top of experiment
	if err := store.GetDirectory(path.Join("commits", com.ID), outputDir); err != nil {
		return err
	}
	fmt.Println()
	console.Info("Checked out %s to %q", com.ShortID(), outputDir)
	return nil
}
