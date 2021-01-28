package project

import (
	"fmt"
	"path/filepath"

	"github.com/replicate/keepsake/go/pkg/console"
	"github.com/replicate/keepsake/go/pkg/errors"
)

func (p *Project) CheckoutCheckpoint(checkpoint *Checkpoint, experiment *Experiment, outputDir string, quiet bool) error {
	// TODO: This function checks out both experiments and checkpoints. This logic should probably be split out so those two things can be done explicitly. This will involve moving some logic to cli/checkpoint.go

	if checkpoint == nil {
		if experiment.Path == "" {
			return errors.DoesNotExist(fmt.Sprintf("The experiment %s does not have any files associated with it. You need to pass the 'path' argument to 'init()' to check out files.", experiment.ShortID()))
		}
	} else {
		if experiment.Path == "" && checkpoint.Path == "" {
			return errors.DoesNotExist(fmt.Sprintf("Neither the checkpoint %s nor its experiment experiment %s have any files associated with them. You need to pass the 'path' argument to 'init()' or 'checkpoint()' to check out files.", checkpoint.ShortID(), experiment.ShortID()))
		}
	}

	if experiment.Path != "" {
		if !quiet {
			console.Info("Copying files from experiment %s to %q...", experiment.ShortID(), filepath.Join(outputDir, experiment.Path))
		}
		if err := p.repository.GetPathTar(experiment.StorageTarPath(), outputDir); err != nil {
			if errors.IsDoesNotExist(err) {
				return errors.DoesNotExist(fmt.Sprintf("Experiment %s is supposed to have files associated with it, but could not find the files at %q.\nMaybe it hasn't been written yet, or the repository is corrupted?", experiment.ShortID(), experiment.StorageTarPath()))
			} else {
				return err
			}
		}
	}

	// Overlay checkpoint on top of experiment
	if checkpoint != nil && checkpoint.Path != "" {
		if !quiet {
			console.Info("Copying files from checkpoint %s to %q...", checkpoint.ShortID(), filepath.Join(outputDir, checkpoint.Path))
		}

		if err := p.repository.GetPathTar(checkpoint.StorageTarPath(), outputDir); err != nil {
			if errors.IsDoesNotExist(err) {
				return errors.DoesNotExist(fmt.Sprintf("Checkpoint %s is supposed to have files associated with it, but could not find the files at %q.\nMaybe it hasn't been written yet, or the repository is corrupted?", checkpoint.ShortID(), checkpoint.StorageTarPath()))
			} else {
				return err

			}
		}
	}

	if !quiet {
		console.Info(`If you want to run this experiment again, this is how it was run:

  ` + experiment.Command + `
`)
	}

	return nil
}

// checkout all the files from an experiment or checkpoint
func (p *Project) CheckoutFileOrDirectory(checkpoint *Checkpoint, experiment *Experiment, outputDir string, checkoutPath string) error {
	// Extract the tarfile
	experimentFilesExist := true
	checkpointFilesExist := true

	if err := p.repository.GetPathItemTar(filepath.Join("experiments", experiment.ID+".tar.gz"), checkoutPath, outputDir); err != nil {
		// Ignore does not exist errors
		if errors.IsDoesNotExist(err) {
			console.Debug("No experiment data found")
			experimentFilesExist = false
		} else {
			return err
		}
	} else {
		console.Info("Copied the path %s from experiment %s to %q", checkoutPath, experiment.ShortID(), filepath.Join(outputDir, experiment.Path))
	}

	// Overlay checkpoint on top of experiment
	if checkpoint != nil {

		if err := p.repository.GetPathItemTar(filepath.Join("checkpoints", checkpoint.ID+".tar.gz"), checkoutPath, outputDir); err != nil {
			if errors.IsDoesNotExist(err) {
				console.Debug("No checkpoint data found")
				checkpointFilesExist = false
			} else {
				return err

			}
		} else {
			console.Info("Copied the path %s from checkpoint %s to %q", checkoutPath, checkpoint.ShortID(), filepath.Join(outputDir, checkpoint.Path))
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
