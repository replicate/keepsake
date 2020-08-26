package cli

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"strings"

	"github.com/docker/docker/client"
	"github.com/otiai10/copy"
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/build"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/docker"
	"replicate.ai/cli/pkg/global"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/netutils"
	"replicate.ai/cli/pkg/remote"
	"replicate.ai/cli/pkg/settings"
	"replicate.ai/cli/pkg/storage"
)

type runOpts struct {
	host       string
	privateKey string
	mounts     []string
}

func newRunCommand() *cobra.Command {
	var opts runOpts

	cmd := &cobra.Command{
		Use:   "run [flags] <command> [arg...]",
		Short: "Run a command on a remote machine",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCommand(opts, args)
		},
	}

	flags := cmd.Flags()
	// Flags after first argument are interpreted as arguments, so they get passed to Docker
	flags.SetInterspersed(false)
	flags.StringVarP(&opts.host, "host", "H", "", "SSH host to run command on, in form [username@]hostname[:port]")
	flags.StringVarP(&opts.privateKey, "identity-file", "i", "", "SSH private key path")

	// TODO(andreas): mounts really ought to be defined in replicate.yaml since models probably wont work without them existing
	flags.StringArrayVarP(&opts.mounts, "mount", "m", []string{}, "Mount directories from the host to Replicate's Docker container. Format: --mount <host-directory>:<container-directory>")

	return cmd
}

func runCommand(opts runOpts, args []string) (err error) {
	conf, sourceDir, err := loadConfig()
	if err != nil {
		return err
	}
	console.Debug("Using directory: %s", sourceDir)

	// User input checks
	if opts.host != "" {
		scheme, _, err := storage.SplitURL(conf.Storage)
		if err != nil {
			return err
		}
		if scheme == storage.SchemeDisk {
			return fmt.Errorf("You can't run commands on remote machines with local disk storage.\n\nSet the 'storage' option in replicate.yaml to an s3:// or gs:// URL.\n\nSee the docs for more information: " + global.WebURL + "/docs/working-with-remote-machines")
		}
	}

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

	hostCUDADriverVersion := ""
	if remoteOptions != nil {
		remoteClient, err := remote.NewClient(remoteOptions)
		if err != nil {
			return fmt.Errorf("Error creating remote client: %w", err)
		}
		hostCUDADriverVersion, err = remoteClient.GetCUDADriverVersion()
		if err != nil {
			return fmt.Errorf("Error getting CUDA driver version: %w", err)
		}

		if hostCUDADriverVersion == "" {
			console.Info("No CUDA driver found on remote host")
		} else {
			console.Info("Found CUDA driver on remote host: %s", hostCUDADriverVersion)
		}
	}
	hasGPU := hostCUDADriverVersion != ""

	console.Info("Building Docker image...")
	baseImage, err := build.GetBaseImage(conf, sourceDir, hostCUDADriverVersion)
	if err != nil {
		return err
	}

	// In development, put development Python package inside Docker image
	devPythonSource := os.Getenv("REPLICATE_DEV_PYTHON_SOURCE")
	devPythonSourceTmpdir := ""
	if devPythonSource != "" {
		console.Info("Using Python library from %s", devPythonSource)
		// Python source needs to be inside source dir so it gets uploaded and is
		// in build context. It also can't be inside anything that doesn't
		// get rsynced (e.g. .replicate)
		devPythonSourceTmpdir = ".tmp-dev-python-source"
		if err := os.RemoveAll(path.Join(sourceDir, devPythonSourceTmpdir)); err != nil {
			return err
		}
		if err := copy.Copy(devPythonSource, path.Join(sourceDir, devPythonSourceTmpdir)); err != nil {
			return fmt.Errorf("Failed to copy REPLICATE_DEV_PYTHON_SOURCE: %w", err)
		}
	}

	console.Debug("Using base image: %s", baseImage.RepositoryName())
	dockerfile, err := build.GenerateDockerfile(conf, sourceDir, devPythonSourceTmpdir)
	if err != nil {
		return err
	}

	if err := docker.Build(remoteOptions, sourceDir, dockerfile, containerName, baseImage.RepositoryName(), hasGPU); err != nil {
		return err
	}

	// Add a bit of space
	fmt.Println()

	// Prepend `python` for convenience
	if strings.HasSuffix(args[0], ".py") {
		args = append([]string{"python", "-u"}, args...)
	}

	// forward the local username (using environment variable)
	// to the container, which will get saved in metadata
	username, err := getUser()
	if err != nil {
		return err
	}

	// also forward the host: if we're running with --host,
	// use that. otherwise use the local outbound IP
	var host string
	if remoteOptions == nil {
		host, err = netutils.GetOutboundIP()
		if err != nil {
			return err
		}
	} else {
		host = remoteOptions.Host
	}

	mounts, err := parseMounts(opts.mounts)
	if err != nil {
		return err
	}

	console.Info("Running '%v'...", strings.Join(args, " "))
	return docker.Run(dockerClient, containerName, args, mounts, hasGPU, username, host, conf.Storage)
}

func parseMounts(mountStrings []string) ([]docker.Mount, error) {
	mounts := []docker.Mount{}
	for _, s := range mountStrings {
		parts := strings.Split(s, ":")
		if len(parts) != 2 {
			return nil, fmt.Errorf("Mount is not in the format \"<host-directory>:<container-directory>\": %s", s)
		}
		mounts = append(mounts, docker.Mount{
			HostDir:      parts[0],
			ContainerDir: parts[1],
		})
	}
	return mounts, nil
}

func getUser() (string, error) {
	userSettings, err := settings.LoadUserSettings()
	if err != nil || userSettings.Email == "" {
		u, err := user.Current()
		if err != nil {
			return "", fmt.Errorf("Failed to determine current user, got error: %w", err)
		}
		return u.Username, nil
	}
	return userSettings.Email, nil
}
