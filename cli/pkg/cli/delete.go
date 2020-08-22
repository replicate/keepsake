package cli

import (
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/project"
	"replicate.ai/cli/pkg/storage"
)

func newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <experiment-or-commit-id> [experiment-or-commit-id...]",
		Short: "Delete experiments or commits",
		Long: `Delete experiments or commits.

To delete experiments or commits, pass any number of IDs (or prefixes).
`,
		RunE: deleteExperimentOrCommit,
		Args: cobra.MinimumNArgs(1),
	}

	addStorageURLFlag(cmd)

	return cmd
}

func deleteExperimentOrCommit(cmd *cobra.Command, prefixes []string) error {
	storageURL, _, err := getStorageURLFromFlagOrConfig(cmd)
	if err != nil {
		return err
	}
	store, err := storage.ForURL(storageURL)
	if err != nil {
		return err
	}
	proj := project.NewProject(store)
	if err != nil {
		return err
	}

	for _, prefix := range prefixes {
		comOrExp, err := proj.CommitOrExperimentFromPrefix(prefix)
		if err != nil {
			return err
		}
		if comOrExp.Commit != nil {
			console.Info("Deleting commit %s...", comOrExp.Commit.ID)
			if err := proj.DeleteCommit(comOrExp.Commit); err != nil {
				return err
			}
		} else {
			console.Info("Deleting experiment %s and its commits...", comOrExp.Experiment.ID)
			commits, err := proj.ExperimentCommits(comOrExp.Experiment.ID)
			if err != nil {
				return err
			}
			for _, commit := range commits {
				if err := proj.DeleteCommit(commit); err != nil {
					return err
				}
			}
			if err := proj.DeleteExperiment(comOrExp.Experiment); err != nil {
				return err
			}
		}
	}

	return nil
}
