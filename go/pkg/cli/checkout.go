package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/replicate/keepsake/go/pkg/console"
	"github.com/replicate/keepsake/go/pkg/files"
	"github.com/replicate/keepsake/go/pkg/project"
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
	cmd.Flags().StringVarP(&opts.outputDirectory, "output-directory", "o", "", "Output directory (defaults to working directory or directory with keepsake.yaml in it)")
	cmd.Flags().BoolVarP(&opts.force, "force", "f", false, "Force checkout without prompt, even if the directory is not empty")
	cmd.Flags().StringVarP(&opts.checkoutPath, "path", "", "", "A specific file or directory to checkout (defaults to all files or directory in checkpoint/experiment)")

	return cmd
}

// Returns the experiment and the most appropriate checkpoint for that experiment.
func getExperimentAndCheckpoint(prefix string, proj *project.Project, projectDir string) (*project.Experiment, *project.Checkpoint, error) {
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
func validateOrCreateOutputDir(outputDir string) error {
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
		isDir, err := files.IsDir(displayPath)
		if err != nil {
			return err
		}
		if isDir {
			isEmpty, err := files.DirIsEmpty(displayPath)
			if err != nil {
				return err
			}
			if !isEmpty && !force {
				console.Warn("The directory %q is not empty.", displayPath)
				console.Warn("%s Make sure they're saved in Git or Keepsake so they're safe!", aurora.Bold("This checkout may overwrite existing files."))
				fmt.Println()
				// This is scary! See https://github.com/replicate/keepsake/issues/300
				doOverwrite, err := console.InteractiveBool{
					Prompt:         "Do you want to continue?",
					Default:        false,
					NonDefaultFlag: "-f",
				}.Read()
				if err != nil {
					return err
				}
				if !doOverwrite {
					return fmt.Errorf("Aborting.")
				}
			}
		} else if !force {
			// it's a file
			console.Warn("The file %q exists.", displayPath)
			console.Warn("%s Make sure it's saved in Git or Keepsake so it's safe!", aurora.Bold("This checkout may overwrite existing files."))
			fmt.Println()
			// This is scary! See https://github.com/replicate/keepsake/issues/300
			doOverwrite, err := console.InteractiveBool{
				Prompt:         "Do you want to continue?",
				Default:        false,
				NonDefaultFlag: "-f",
			}.Read()
			if err != nil {
				return err
			}
			if !doOverwrite {
				return fmt.Errorf("Aborting.")
			}
		}
	}
	return nil
}

// keepsake CLI `checkout` command
func checkoutCheckpoint(opts checkoutOpts, args []string) error {
	prefix := args[0]

	repositoryURL, projectDir, err := getRepositoryURLFromStringOrConfig(opts.repositoryURL)
	if err != nil {
		return err
	}
	repo, err := getRepository(repositoryURL, projectDir)
	if err != nil {
		return err
	}

	proj := project.NewProject(repo, projectDir)
	experiment, checkpoint, err := getExperimentAndCheckpoint(prefix, proj, projectDir)
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

	err = validateOrCreateOutputDir(outputDir)
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
		return proj.CheckoutCheckpoint(checkpoint, experiment, outputDir, false)
	} else {
		return proj.CheckoutFileOrDirectory(checkpoint, experiment, outputDir, checkoutPath)
	}
}
