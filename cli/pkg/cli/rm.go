package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/interact"
	"replicate.ai/cli/pkg/project"
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
	}

	addStorageURLFlag(cmd)
	cmd.Flags().BoolP("force", "f", false, "Force delete without interactive prompt")

	return cmd
}

func removeExperimentOrCheckpoint(cmd *cobra.Command, prefixes []string) error {
	storageURL, projectDir, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	store, err := getStorage(storageURL, projectDir)
	if err != nil {
		return err
	}
	proj := project.NewProject(store)
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
