package cli

import (
	"github.com/spf13/cobra"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/global"
	"github.com/replicate/replicate/go/pkg/project"
	"github.com/replicate/replicate/go/pkg/shared"
)

func NewDaemonCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "replicate-daemon <socket-path>",
		RunE: runDaemon,
	}
	setPersistentFlags(cmd)
	addRepositoryURLFlag(cmd)
	return cmd
}

func runDaemon(cmd *cobra.Command, args []string) error {
	socketPath := args[0]

	if global.Verbose {
		console.SetLevel(console.DebugLevel)
	}

	projectGetter := func() (proj *project.Project, err error) {
		repositoryURL, projectDir, err := getRepositoryURLFromFlagOrConfig(cmd)
		if err != nil {
			return nil, err
		}
		repo, err := getRepository(repositoryURL, projectDir)
		if err != nil {
			return nil, err
		}
		proj = project.NewProject(repo, projectDir)
		return proj, err
	}

	if err := shared.Serve(projectGetter, socketPath); err != nil {
		return err
	}
	return nil
}
