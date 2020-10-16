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
		Use:   "checkout <experiment or checkpoint ID>",
		Short: "Copy files from an experiment or checkpoint into the project directory",
		Run:   handleErrors(checkoutCheckpoint),
		Args:  cobra.ExactArgs(1),
	}

	addStorageURLFlag(cmd)
	cmd.Flags().StringP("output-directory", "o", "", "Output directory (defaults to working directory or directory with replicate.yaml in it)")
	cmd.Flags().BoolP("force", "f", false, "Force checkout without prompt, even if the directory is not empty")

	return cmd
}

func checkoutCheckpoint(cmd *cobra.Command, args []string) error {
	prefix := args[0]

	outputDir, err := cmd.Flags().GetString("output-directory")
	if err != nil {
		return err
	}
	// TODO(andreas): add test for case where --output-directory is omitted
	if outputDir == "" {
		outputDir, err = getProjectDir()
		if err != nil {
			return err
		}
	}

	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	storageURL, projectDir, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	store, err := getStorage(storageURL, projectDir)
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
			return fmt.Errorf("Failed to create directory %q: %w", outputDir, err)
		}
	}

	proj := project.NewProject(store)
	result, err := proj.CheckpointOrExperimentFromPrefix(prefix)
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

	msg := ""
	experiment := result.Experiment

	if result.Checkpoint != nil {
		// Checking out checkpoint
		checkpoint := result.Checkpoint

		if err := store.GetPathTar(path.Join("experiments", experiment.ID+".tar.gz"), outputDir); err != nil {
			// Ignore does not exist errors
			if _, ok := err.(*storage.DoesNotExistError); !ok {
				return err
			}
		}
		// Overlay checkpoint on top of experiment
		if err := store.GetPathTar(path.Join("checkpoints", checkpoint.ID+".tar.gz"), outputDir); err != nil {
			if _, ok := err.(*storage.DoesNotExistError); ok {
				return fmt.Errorf(`Could not check out this checkpoint because it does not have any files associated with it.

You need to set the 'path' option on 'checkpoint()' to check them out.`)
			}
			return err
		}

		// TODO: actually mention which files were checked out
		msg += fmt.Sprintf("Copied the files from checkpoint %s to %q\n", checkpoint.ShortID(), outputDir)

	} else {
		// Checking out experiment
		if err := store.GetPathTar(path.Join("experiments", experiment.ID+".tar.gz"), outputDir); err != nil {
			if _, ok := err.(*storage.DoesNotExistError); ok {
				return fmt.Errorf(`Could not check out this experiment because it does not have any files associated with it.

You need to set the 'path' option on 'init()' to check experiments out.`)
			}
			return err
		}

		msg += fmt.Sprintf("Copied the files from experiment %s to %q\n", experiment.ShortID(), outputDir)
	}

	msg += "\n"
	msg += "If you want to run this experiment again, this is how it was run:\n"
	msg += "\n"
	msg += "  " + experiment.Command
	msg += "\n"
	fmt.Fprintln(os.Stderr)
	console.Info(msg)

	return nil
}
