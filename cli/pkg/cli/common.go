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
func getStorageURLFromFlagOrConfig(cmd *cobra.Command) (storageURL string, sourceDir string, err error) {
	storageURL, err = cmd.Flags().GetString("storage-url")
	if err != nil {
		return "", "", err
	}

	if storageURL == "" {
		conf, sourceDir, err := config.FindConfigInWorkingDir(global.SourceDirectory)
		if err != nil {
			return "", "", err
		}
		return conf.Storage, sourceDir, nil
	}

	// if global.SourceDirectory == "", abs of that is cwd
	// FIXME (bfirsh): this does not look up directories for replicate.yaml, so might be the wrong
	// sourceDir. It should probably use return value of FindConfigInWorkingDir.
	sourceDir, err = filepath.Abs(global.SourceDirectory)
	if err != nil {
		return "", "", fmt.Errorf("Failed to determine absolute directory of '%s', got error: %w", global.SourceDirectory, err)
	}

	return storageURL, sourceDir, nil
}

// getSourceDir returns the project's source directory
func getSourceDir() (string, error) {
	_, sourceDir, err := config.FindConfigInWorkingDir(global.SourceDirectory)
	if err != nil {
		return "", err
	}
	return sourceDir, nil
}

// getStorage returns the project's storage, with caching if needed
// This is not in storage package so we can do user interface stuff around syncing
func getStorage(storageURL, sourceDir string) (storage.Storage, error) {
	store, err := storage.ForURL(storageURL)
	if err != nil {
		return nil, err
	}
	// sourceDir might be "" if you use --storage-url option
	if storage.NeedsCaching(store) && sourceDir != "" {
		console.Info("Fetching new data from %q...", store.RootURL())
		store, err = storage.NewCachedMetadataStorage(store, sourceDir)
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
