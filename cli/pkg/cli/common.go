package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/global"
	"replicate.ai/cli/pkg/storage"
)

func getAurora() aurora.Aurora {
	// TODO (bfirsh): consolidate this logic in console package
	return aurora.NewAurora(os.Getenv("NO_COLOR") == "")
}

func addStorageURLFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("storage-url", "S", "", "Storage URL (e.g. 's3://my-replicate-bucket' (if omitted, uses storage URL from replicate.yaml)")
}

// getStorageURLFromConfigOrFlag uses --storage-url if it exists,
// otherwise finds replicate.yaml recursively
func getStorageURLFromFlagOrConfig(cmd *cobra.Command) (storageURL string, projectDir string, err error) {
	storageURL, err = cmd.Flags().GetString("storage-url")
	if err != nil {
		return "", "", err
	}

	if storageURL == "" {
		conf, projectDir, err := config.FindConfigInWorkingDir(global.ProjectDirectory)
		if err != nil {
			return "", "", err
		}
		return conf.Storage, projectDir, nil
	}

	// if global.ProjectDirectory == "", abs of that is cwd
	// FIXME (bfirsh): this does not look up directories for replicate.yaml, so might be the wrong
	// projectDir. It should probably use return value of FindConfigInWorkingDir.
	projectDir, err = filepath.Abs(global.ProjectDirectory)
	if err != nil {
		return "", "", fmt.Errorf("Failed to determine absolute directory of '%s': %w", global.ProjectDirectory, err)
	}

	return storageURL, projectDir, nil
}

// getProjectDir returns the project's source directory
func getProjectDir() (string, error) {
	_, projectDir, err := config.FindConfigInWorkingDir(global.ProjectDirectory)
	if err != nil {
		return "", err
	}
	return projectDir, nil
}

// getStorage returns the project's storage, with caching if needed
// This is not in storage package so we can do user interface stuff around syncing
func getStorage(storageURL, projectDir string) (storage.Storage, error) {
	store, err := storage.ForURL(storageURL)
	if err != nil {
		return nil, err
	}
	// projectDir might be "" if you use --storage-url option
	if storage.NeedsCaching(store) && projectDir != "" {
		console.Info("Fetching new data from %q...", store.RootURL())
		store, err = storage.NewCachedMetadataStorage(store, projectDir)
		if err != nil {
			return nil, err
		}
		cachedStore := store.(*storage.CachedStorage)
		if err := cachedStore.SyncCache(); err != nil {
			return nil, err
		}
	}
	return store, nil
}

// handlErrors wraps a cobra function, and will print and exit on error
//
// We don't use RunE because if that returns an error, Cobra will print usage.
// That behavior can be disabled with SilenceUsage option, but then Cobra arg/flag errors don't display usage. (sigh)
func handleErrors(f func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		if err := f(cmd, args); err != nil {
			console.Fatal(err.Error())
		}
	}
}
