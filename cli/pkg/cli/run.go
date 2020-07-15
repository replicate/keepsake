package cli

import (
	"fmt"
	"strings"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/build"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/docker"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/remote"
)

type runOpts struct {
	host       string
	privateKey string
}

func newRunCommand() *cobra.Command {
	var opts runOpts

	cmd := &cobra.Command{
		Use:   "run [OPTIONS] COMMAND [ARG...]",
		Short: "Run a command on a remote machine",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, args)
		},
	}

	flags := cmd.Flags()
	// Flags after first argument are interpreted as arguments, so they get passed to Docker
	flags.SetInterspersed(false)
	flags.StringVarP(&opts.host, "host", "H", "", "SSH host to run command on, in form [username@]hostname[:port]")
	flags.StringVarP(&opts.privateKey, "identity-file", "i", "", "SSH private key path")

	return cmd
}

func runCommand(opts runOpts, args []string) (err error) {
	var remoteOptions *remote.Options
	var dockerClient *client.Client

	if opts.host == "" {
		// Local mode
		dockerClient, err = docker.NewLocalClient()
		if err != nil {
			return err
		}
	} else {
		// Remote SSH mode
		remoteOptions, err = remote.ParseHost(opts.host)
		if err != nil {
			return err
		}
		if opts.privateKey != "" {
			remoteOptions.PrivateKeys = []string{opts.privateKey}
		}

		dockerClient, err = docker.NewRemoteClient(remoteOptions)
		if err != nil {
			return err
		}
	}

	// TODO: maybe make this same as experiment ID? could generate environment ID here and pass as environment variable
	// to Python library or something.
	jobID := hash.Random()
	containerName := "replicate-" + jobID

	conf, sourceDir, err := loadConfig()
	if err != nil {
		return err
	}
	console.Debug("Using directory: %s", sourceDir)
	console.Info("Building Docker image...")

	hostCUDADriverVersion := ""
	if remoteOptions != nil {
		remoteClient, err := remote.NewClient(remoteOptions)
		if err != nil {
			return err
		}
		hostCUDADriverVersion, err = remoteClient.GetCUDADriverVersion()
		if err != nil {
			return err
		}

		if hostCUDADriverVersion == "" {
			console.Debug("No CUDA driver found on remote host")
		} else {
			console.Debug("Found CUDA driver on remote host: %s", hostCUDADriverVersion)
		}
	}
	hasGPU := hostCUDADriverVersion != ""

	baseImage, err := build.GetBaseImage(conf, sourceDir, hostCUDADriverVersion)
	if err != nil {
		return err
	}
	dockerfile, err := build.GenerateDockerfile(conf, sourceDir)
	if err != nil {
		return err
	}

	if err := docker.Build(remoteOptions, sourceDir, dockerfile, containerName, baseImage.RepositoryName(), hasGPU); err != nil {
		return err
	}

	// Add a bit of space
	fmt.Println()

	console.Info("Running '%v'...", strings.Join(args, " "))
	return docker.Run(dockerClient, containerName, args)
}
