package project

import (
	"os"
	"path"
	"testing"
	"time"

	"github.com/replicate/replicate/go/pkg/config"
	"github.com/replicate/replicate/go/pkg/files"
	"github.com/replicate/replicate/go/pkg/repository"
	"github.com/stretchr/testify/require"
)

func TestCheckoutWithNoPaths(t *testing.T) {
	projectDir, err := files.TempDir("test-checkout")
	require.NoError(t, err)
	defer os.RemoveAll(projectDir)

	repo, err := repository.NewDiskRepository(path.Join(projectDir, ".replicate"))
	require.NoError(t, err)

	fixedTime, _ := time.Parse(time.RFC3339, "2006-01-02T15:04:05Z")
	require.NoError(t, err)

	experiment := &Experiment{
		ID:      "1eeeeeeeee",
		Created: fixedTime.Add(-10 * time.Minute),
		Config:  &config.Config{},
		Path:    "",
		Checkpoints: []*Checkpoint{
			{
				ID:      "2ccccccccc",
				Created: fixedTime.Add(-5 * time.Minute),
				Path:    "",
			},
		},
	}
	require.NoError(t, experiment.Save(repo))

	project := NewProject(repo, projectDir)

	err = project.CheckoutCheckpoint(experiment.Checkpoints[0], experiment, projectDir, true)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Neither the checkpoint 2cccccc nor its experiment experiment 1eeeeee have any files associated with them.")

	err = project.CheckoutCheckpoint(nil, experiment, projectDir, true)
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "The experiment 1eeeeee does not have any files associated with it.")
}
