package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/interact"
	"github.com/replicate/replicate/go/pkg/project"
)

func newRmCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rm <experiment or checkpoint ID> [experiment or checkpoint ID...]",
		Short: "Remove experiments or checkpoint",
		Long: `Remove experiments or checkpoints.

To remove experiments or checkpoints, pass any number of IDs (or prefixes).
`,
		Run:        handleErrors(removeExperimentOrCheckpoint),
		Args:       cobra.MinimumNArgs(1),
		Aliases:    []string{"delete"},
		SuggestFor: []string{"remove"},
		Example: `Delete an experiment and its checkpoints
(where a1b2c3d4 is an experiment ID):
replicate rm a1b2c3d4

Delete all experiments where the metric "val_accuracy" is less
than 0.2 at the best checkpoints:
replicate rm $(replicate ls -q --filter "val_accuracy < 0.2")
`,
	}

	addRepositoryURLFlag(cmd)
	cmd.Flags().BoolP("force", "f", false, "Force delete without interactive prompt")

	return cmd
}

func removeExperimentOrCheckpoint(cmd *cobra.Command, prefixes []string) error {
	repositoryURL, projectDir, err := getRepositoryURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	repo, err := getRepository(repositoryURL, projectDir)
	if err != nil {
		return err
	}
	proj := project.NewProject(repo)
	if err != nil {
		return err
	}
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return err
	}

	comOrExps := []*project.CheckpointOrExperiment{}
	for _, prefix := range prefixes {
		comOrExp, err := proj.CheckpointOrExperimentFromPrefix(prefix)
		if err != nil {
			return err
		}
		comOrExps = append(comOrExps, comOrExp)
	}

	if len(comOrExps) == 0 {
		return nil
	}

	if !force {
		fmt.Println("You are about to delete the following:")
		for _, comOrExp := range comOrExps {
			if comOrExp.Experiment != nil {
				fmt.Printf("* Experiment %s (%d checkpoints)\n", comOrExp.Experiment.ShortID(), len(comOrExp.Experiment.Checkpoints))
			} else {
				fmt.Printf("* Checkpoint %s\n", comOrExp.Checkpoint.ShortID())
			}
		}
		continueDelete, err := interact.InteractiveBool{
			Prompt:  "\nDo you want to continue?",
			Default: false,
		}.Read()
		if err != nil {
			return err
		}
		if !continueDelete {
			return nil
		}
	}

	for _, prefix := range prefixes {
		comOrExp, err := proj.CheckpointOrExperimentFromPrefix(prefix)
		if err != nil {
			return err
		}
		if comOrExp.Checkpoint != nil {
			console.Info("Removing checkpoint %s...", comOrExp.Checkpoint.ShortID())
			if err := proj.DeleteCheckpoint(comOrExp.Checkpoint); err != nil {
				return err
			}
		} else {
			console.Info("Removing experiment %s and its checkpoints...", comOrExp.Experiment.ShortID())
			experiment := comOrExp.Experiment
			// This is slow, see https://github.com/replicate/replicate/issues/333
			for _, checkpoint := range experiment.Checkpoints {
				if err := proj.DeleteCheckpoint(checkpoint); err != nil {
					return err
				}
			}
			if err := proj.DeleteExperiment(experiment); err != nil {
				return err
			}
		}
	}

	return nil
}
