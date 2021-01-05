package project

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/errors"
)

func (p *Project) CheckoutCheckpoint(checkpoint *Checkpoint, experiment *Experiment, outputDir string, quiet bool) error {
	experimentFilesExist := true
	checkpointFilesExist := true

	if err := p.repository.GetPathTar(path.Join("experiments", experiment.ID+".tar.gz"), outputDir); err != nil {
		// Ignore does not exist errors
		if errors.IsDoesNotExist(err) {
			console.Debug("No experiment data found")
			experimentFilesExist = false
		} else {
			return err
		}
	} else {
		if !quiet {
			console.Info("Copied the files from experiment %s to %q", experiment.ShortID(), filepath.Join(outputDir, experiment.Path))
		}
	}

	// Overlay checkpoint on top of experiment
	if checkpoint != nil {

		if err := p.repository.GetPathTar(path.Join("checkpoints", checkpoint.ID+".tar.gz"), outputDir); err != nil {
			if errors.IsDoesNotExist(err) {
				console.Debug("No checkpoint data found")
				checkpointFilesExist = false
			} else {
				return err

			}
		} else {
			if !quiet {
				console.Info("Copied the files from checkpoint %s to %q", checkpoint.ShortID(), filepath.Join(outputDir, checkpoint.Path))
			}
		}

	}

	if !experimentFilesExist && !checkpointFilesExist {
		// Just an experiment, no checkpoints
		if checkpoint == nil {
			return fmt.Errorf("The experiment %s does not have any files associated with it. You need to pass the 'path' argument to 'init()' to check out files.", experiment.ShortID())
		}
		return errors.DoesNotExist(fmt.Sprintf("Neither the experiment %s nor the checkpoint %s has any files associated with it. You need to pass the 'path' argument to 'init()' or 'checkpoint()' to check out files.", experiment.ShortID(), checkpoint.ShortID()))
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
