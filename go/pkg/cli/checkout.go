package cli

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/files"
	"github.com/replicate/replicate/go/pkg/interact"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/repository"
)

func newCheckoutCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "checkout <experiment or checkpoint ID>",
		Short: "Copy files from an experiment or checkpoint into the project directory",
		Run:   handleErrors(checkoutCheckpoint),
		Args:  cobra.ExactArgs(1),
	}

	addRepositoryURLFlag(cmd)
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

	repositoryURL, projectDir, err := getRepositoryURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	repo, err := getRepository(repositoryURL, projectDir)
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

	proj := project.NewProject(repo)
	result, err := proj.CheckpointOrExperimentFromPrefix(prefix)
	if err != nil {
		return err
	}

	experiment := result.Experiment
	checkpoint := result.Checkpoint

	if checkpoint != nil {
		console.Info("Checking out files from checkpoint %s and its experiment %s", checkpoint.ShortID(), experiment.ShortID())
	} else {
		// When checking out experiment, also check out best/latest checkpoint
		checkpoint = experiment.BestCheckpoint()
		if checkpoint != nil {
			console.Info("Checking out files from experiment %s and its best checkpoint %s", experiment.ShortID(), checkpoint.ShortID())
		} else {
			checkpoint = experiment.LatestCheckpoint()
			if checkpoint != nil {
				console.Info("Checking out files from experiment %s and its latest checkpoint %s", experiment.ShortID(), checkpoint.ShortID())
			} else {
				console.Info("Checking out files from experiment %s", experiment.ShortID())
			}
		}
	}

	displayPath := filepath.Join(outputDir, experiment.Path)

	// FIXME(bfirsh): this is a bodge and isn't always quite right -- if no experiment path set, and we're checking out checkpoint, display the checkpoint path
	if experiment.Path == "" && checkpoint != nil {
		displayPath = filepath.Join(outputDir, checkpoint.Path)
	}

	isEmpty, err := files.DirIsEmpty(displayPath)
	if err != nil {
		return err
	}
	if !isEmpty && !force {
		console.Warn("The directory %q is not empty.", displayPath)
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

	fmt.Fprintln(os.Stderr)

	experimentFilesExist := true
	checkpointFilesExist := true

	if err := repo.GetPathTar(path.Join("experiments", experiment.ID+".tar.gz"), outputDir); err != nil {
		// Ignore does not exist errors
		if _, ok := err.(*repository.DoesNotExistError); ok {
			console.Debug("No experiment data found")
			experimentFilesExist = false
		} else {
			return err
		}
	} else {
		console.Info("Copied the files from experiment %s to %q", experiment.ShortID(), filepath.Join(outputDir, experiment.Path))
	}

	// Overlay checkpoint on top of experiment
	if checkpoint != nil {

		if err := repo.GetPathTar(path.Join("checkpoints", checkpoint.ID+".tar.gz"), outputDir); err != nil {
			if _, ok := err.(*repository.DoesNotExistError); ok {
				console.Debug("No checkpoint data found")
				checkpointFilesExist = false
			} else {
				return err

			}
		} else {
			console.Info("Copied the files from checkpoint %s to %q", checkpoint.ShortID(), filepath.Join(outputDir, checkpoint.Path))
		}

	}

	if !experimentFilesExist && !checkpointFilesExist {
		// Just an experiment, no checkpoints
		if checkpoint == nil {
			return fmt.Errorf("The experiment %s does not have any files associated with it. You need to pass the 'path' argument to 'init()' to check out files.", experiment.ShortID())
		}
		return fmt.Errorf("Neither the experiment %s nor the checkpoint %s has any files associated with it. You need to pass the 'path' argument to 'init()' or 'checkpoint()' to check out files.", experiment.ShortID(), checkpoint.ShortID())
	}

	console.Info(`If you want to run this experiment again, this is how it was run:

  ` + experiment.Command + `
`)

	return nil
}
