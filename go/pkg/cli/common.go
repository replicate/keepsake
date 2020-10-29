package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/logrusorgru/aurora"
	"github.com/spf13/cobra"

	"github.com/replicate/replicate/go/pkg/config"
	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/global"
	"github.com/replicate/replicate/go/pkg/repository"
)

func getAurora() aurora.Aurora {
	// TODO (bfirsh): consolidate this logic in console package
	return aurora.NewAurora(os.Getenv("NO_COLOR") == "")
}

func addRepositoryURLFlag(cmd *cobra.Command) {
	cmd.Flags().StringP("repository", "R", "", "Repository URL (e.g. 's3://my-replicate-bucket' (if omitted, uses repository URL from replicate.yaml)")
}

func addRepositoryURLFlagVar(cmd *cobra.Command, opt *string) {
	cmd.Flags().StringVarP(opt, "repository", "R", "", "Repository URL (e.g. 's3://my-replicate-bucket' (if omitted, uses repository URL from replicate.yaml)")
}

// getRepositoryURLFromStringOrConfig attempts to get it from passed string from --repository,
// otherwise finds replicate.yaml recursively
func getRepositoryURLFromStringOrConfig(repositoryURL string) (string, string, error) {
	if repositoryURL == "" {
		conf, projectDir, err := config.FindConfigInWorkingDir(global.ProjectDirectory)
		if err != nil {
			return "", "", err
		}
		return conf.Repository, projectDir, nil
	}

	// if global.ProjectDirectory == "", abs of that is cwd
	// FIXME (bfirsh): this does not look up directories for replicate.yaml, so might be the wrong
	// projectDir. It should probably use return value of FindConfigInWorkingDir.
	projectDir, err := filepath.Abs(global.ProjectDirectory)
	if err != nil {
		return "", "", fmt.Errorf("Failed to determine absolute directory of '%s': %w", global.ProjectDirectory, err)
	}

	return repositoryURL, projectDir, nil
}

// getRepositoryURLFromConfigOrFlag uses --repository if it exists,
// otherwise finds replicate.yaml recursively
func getRepositoryURLFromFlagOrConfig(cmd *cobra.Command) (repositoryURL string, projectDir string, err error) {
	repositoryURL, err = cmd.Flags().GetString("repository")
	if err != nil {
		return "", "", err
	}
	return getRepositoryURLFromStringOrConfig(repositoryURL)
}

// getProjectDir returns the project's source directory
func getProjectDir() (string, error) {
	_, projectDir, err := config.FindConfigInWorkingDir(global.ProjectDirectory)
	if err != nil {
		return "", err
	}
	return projectDir, nil
}

// getRepository returns the project's repository, with caching if needed
// This is not in repository package so we can do user interface stuff around syncing
func getRepository(repositoryURL, projectDir string) (repository.Repository, error) {
	repo, err := repository.ForURL(repositoryURL)
	if err != nil {
		return nil, err
	}
	// projectDir might be "" if you use --repository option
	if repository.NeedsCaching(repo) && projectDir != "" {
		console.Info("Fetching new data from %q...", repo.RootURL())
		repo, err = repository.NewCachedMetadataRepository(repo, projectDir)
		if err != nil {
			return nil, err
		}
		cachedRepo := repo.(*repository.CachedRepository)
		if err := cachedRepo.SyncCache(); err != nil {
			return nil, err
		}
	}
	return repo, nil
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
