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
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/repository"
)

type checkoutOpts struct {
	outputDirectory string
	force           bool
	repositoryURL   string
	checkoutPath    string
}

func newCheckoutCommand() *cobra.Command {
	var opts checkoutOpts

	cmd := &cobra.Command{
		Use:   "checkout <experiment or checkpoint ID>",
		Short: "Copy files from an experiment or checkpoint into the project directory",
		Run: handleErrors(func(cmd *cobra.Command, args []string) error {
			return checkoutCheckpoint(opts, args)
		}),
		Args: cobra.ExactArgs(1),
	}

	addRepositoryURLFlagVar(cmd, &opts.repositoryURL)
	cmd.Flags().StringVarP(&opts.outputDirectory, "output-directory", "o", "", "Output directory (defaults to working directory or directory with replicate.yaml in it)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force checkout without prompt, even if the directory is not empty")
	cmd.Flags().StringVarP(&opts.checkoutPath, "path", "", "", "A specific file or directory to checkout (defaults to all files or directory in checkpoint/experiment)")

	return cmd
}

// Returns the repository requested by opts.repositoryURL
func getRepositoryFromOpts(opts checkoutOpts) (repository.Repository, error) {
	repositoryURL, projectDir, err := getRepositoryURLFromStringOrConfig(opts.repositoryURL)
	if err != nil {
		return nil, err
	}

	repo, err := getRepository(repositoryURL, projectDir)
	if err != nil {
		return nil, err
	}

	return repo, nil
}

// Returns the experiment and the most appropriate checkpoint for that experiment.
func getExperimentAndCheckpoint(prefix string, repo repository.Repository) (*project.Experiment, *project.Checkpoint, error) {
	proj := project.NewProject(repo)
	result, err := proj.CheckpointOrExperimentFromPrefix(prefix)
	if err != nil {
		return nil, nil, err
	}
	experiment := result.Experiment
	checkpoint := result.Checkpoint

	if checkpoint != nil {
		console.Info("Checking out files from checkpoint %s and its experiment %s", checkpoint.ShortID(), experiment.ShortID())
		return experiment, checkpoint, nil
	}

	// When checking out experiment, also check out best/latest checkpoint
	checkpoint = experiment.BestCheckpoint()
	if checkpoint != nil {
		console.Info("Checking out files from experiment %s and its best checkpoint %s", experiment.ShortID(), checkpoint.ShortID())
		return experiment, checkpoint, nil
	}

	checkpoint = experiment.LatestCheckpoint()
	if checkpoint != nil {
		console.Info("Checking out files from experiment %s and its latest checkpoint %s", experiment.ShortID(), checkpoint.ShortID())
		return experiment, checkpoint, nil
	}

	console.Info("Checking out files from experiment %s", experiment.ShortID())
	return experiment, checkpoint, nil
}

// Handle errors related to the outputDir
func validateOutputDir(outputDir string) error {
	// FIXME(vastolorde95): If outputPath does not exist, there is no way to distinguish if 
	// it is supposed to be a file or a directory. Should we ask the user for a prompt if 
	// the checkoutPath is a file? This way we will be able to support:
	//
	// replicate checkout abc123 -o out/new_model.pth -file data/model.pth
	exists, err := files.FileExists(outputDir)
	if err != nil {
		return err
	}

	if !exists {
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return fmt.Errorf("Failed to create directory %q: %w", outputDir, err)
		}
	}

	isDir, err := files.IsDir(outputDir)
	if err != nil {
		return err
	}

	if !isDir {
		return fmt.Errorf("Checkout path %q is not a directory", outputDir)
	}
	
	return nil
}

// Prompt user for confirmation if the displayPath already exists
func overwriteDisplayPathPrompt(displayPath string, force bool) error {
	exists, err := files.FileExists(displayPath)
	if err != nil {
		return err
	}

	if exists {
		isEmpty, err := files.DirIsEmpty(displayPath)
		if err != nil {
			return err
		}
		if !isEmpty && !force {
			console.Warn("The directory %q is not empty.", displayPath)
			console.Warn("%s Make sure they're saved in Git or Replicate so they're safe!", aurora.Bold("This checkout may overwrite existing files."))
			fmt.Println()
			// This is scary! See https://github.com/replicate/replicate/issues/300
			doOverwrite, err := console.InteractiveBool{
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
	}
	return nil
}

// replicate CLI `checkout` command
func checkoutCheckpoint(opts checkoutOpts, args []string) error {
	prefix := args[0]

	repo, err := getRepositoryFromOpts(opts)
	if err != nil {
		return err
	}

	experiment, checkpoint, err := getExperimentAndCheckpoint(prefix, repo)
	if err != nil {
		return err
	}

	outputDir := opts.outputDirectory
	if outputDir == "" {
		var err error
		outputDir, err = getProjectDir()
		if err != nil {
			return err
		}
	}

	err = validateOutputDir(outputDir)
	if err != nil {
		return err
	}

	displayPath := filepath.Join(outputDir, experiment.Path)

	// FIXME(bfirsh): this is a bodge and isn't always quite right -- if no experiment path set, and we're checking out checkpoint, display the checkpoint path
	if experiment.Path == "" && checkpoint != nil {
		displayPath = filepath.Join(outputDir, checkpoint.Path)
	}

	err = overwriteDisplayPathPrompt(displayPath, opts.force)
	if err != nil {
		return err
	}

	fmt.Fprintln(os.Stderr)

	checkoutPath := opts.checkoutPath
	if checkoutPath == "" {
		return checkoutEntireCheckpoint(outputDir, repo, experiment, checkpoint)
	} else {
		return checkoutFileOrDir(outputDir, checkoutPath, repo, experiment, checkpoint)
	}
}

// checkout all the files from an experiment or checkpoint
func checkoutEntireCheckpoint(outputDir string, repo repository.Repository, experiment *project.Experiment, checkpoint *project.Checkpoint) error {
	// Extract the tarfile
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

// checkout all the files from an experiment or checkpoint
func checkoutFileOrDir(outputDir string, checkoutPath string, repo repository.Repository, experiment *project.Experiment, checkpoint *project.Checkpoint) error {
	// Extract the tarfile
	experimentFilesExist := true
	checkpointFilesExist := true

	if err := repo.GetPathItemTar(path.Join("experiments", experiment.ID+".tar.gz"), checkoutPath, outputDir); err != nil {
		// Ignore does not exist errors
		if _, ok := err.(*repository.DoesNotExistError); ok {
			console.Debug("No experiment data found")
			experimentFilesExist = false
		} else {
			return err
		}
	} else {
		console.Info("Copied the files %s from experiment %s to %q", checkoutPath, experiment.ShortID(), filepath.Join(outputDir, experiment.Path))
	}

	// Overlay checkpoint on top of experiment
	if checkpoint != nil {

		if err := repo.GetPathItemTar(path.Join("checkpoints", checkpoint.ID+".tar.gz"), checkoutPath, outputDir); err != nil {
			if _, ok := err.(*repository.DoesNotExistError); ok {
				console.Debug("No checkpoint data found")
				checkpointFilesExist = false
			} else {
				return err

			}
		} else {
			console.Info("Copied the files %s from checkpoint %s to %q", checkoutPath, checkpoint.ShortID(), filepath.Join(outputDir, checkpoint.Path))
		}

	}

	if !experimentFilesExist && !checkpointFilesExist {
		// Just an experiment, no checkpoints
		if checkpoint == nil {
			return fmt.Errorf("The experiment %s does not have the path %s associated with it. You need to pass the 'path' argument to 'init()' to check out files.", experiment.ShortID(), checkoutPath)
		}
		return fmt.Errorf("Neither the experiment %s nor the checkpoint %s has the path %s associated with it. You need to pass the 'path' argument to 'init()' or 'checkpoint()' to check out files.", experiment.ShortID(), checkpoint.ShortID(), checkoutPath)
	}

	console.Info(`If you want to run this experiment again, this is how it was run:

  ` + experiment.Command + `
`)

	return nil

}
