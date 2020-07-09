package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	dockerContext "github.com/docker/cli/cli/context/docker"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/spf13/cobra"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/docker"
	"replicate.ai/cli/pkg/global"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/remote"
)

type runOpts struct {
	host       string
	privateKey string
}

type closeFunc func() error

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
	flags.StringVarP(&opts.host, "host", "H", "", "SSH host and port to run command on")
	flags.StringVarP(&opts.privateKey, "identity-file", "i", "", "SSH private key path")

	return cmd
}

// TODO: read python version from replicate.yaml
const dockerfile = `FROM python:3.8
# FIXME: temporary, until this is on pypi or we find a better temporary spot
RUN pip install https://storage.googleapis.com/bfirsh-dev/replicate-python/replicate-0.0.1.tar.gz
COPY . /code
# TODO: cache this properly
RUN [ -f requirements.txt ] && pip install -r requirements.txt || echo 0
WORKDIR /code
`

func runCommand(opts runOpts, args []string) error {
	var err error
	dockerHost := ""
	if opts.host != "" {
		dockerHost = "ssh://" + opts.host
	}

	dockerClient, err := getDockerClient(dockerHost)
	if err != nil {
		return err
	}

	// TODO: maybe make this same as experiment ID? could generate environment ID here and pass as environment variable
	// to Python library or something.
	jobID := hash.Random()
	containerName := "replicate-" + jobID

	sourceDir, err := findSourceDir()
	if err != nil {
		return err
	}

	console.Info("Building Docker image...")

	var remoteOptions *remote.Options
	if opts.host != "" {
		username, host, port, err := parseHost(opts.host)
		if err != nil {
			return err
		}
		remoteOptions = &remote.Options{
			Host:     host,
			Username: username,
			Port:     port,
		}
		if opts.privateKey != "" {
			remoteOptions.PrivateKeys = []string{opts.privateKey}
		}
	}

	if err := docker.Build(remoteOptions, sourceDir, dockerfile, containerName); err != nil {
		return err
	}

	// Add a bit of space
	fmt.Println()

	console.Info("Running '%v'...", strings.Join(args, " "))

	// Options for creating container
	config := &container.Config{
		Image: containerName,
		Cmd:   args,
	}
	// Options for starting container (port bindings, volume bindings, etc)
	hostConfig := &container.HostConfig{
		AutoRemove: false, // TODO: probably true
	}

	ctx, cancelFun := context.WithCancel(context.Background())
	defer cancelFun()

	createResponse, err := dockerClient.ContainerCreate(ctx, config, hostConfig, nil, nil, containerName)
	if err != nil {
		return err
	}
	for _, warning := range createResponse.Warnings {
		console.Warn("WARNING: %s", warning)
	}

	statusChan := waitUntilExit(ctx, dockerClient, createResponse.ID)

	if err := dockerClient.ContainerStart(ctx, createResponse.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	// TODO: detached mode
	var errChan chan error
	close, err := connectToLogs(ctx, dockerClient, &errChan, createResponse.ID)
	if err != nil {
		return err
	}
	defer close()

	if errChan != nil {
		if err := <-errChan; err != nil {
			return err
		}
	}

	status := <-statusChan
	if status != 0 {
		return fmt.Errorf("Command exited with non-zero status code: %v", status)
	}

	return nil
}

func getDockerClient(host string) (*client.Client, error) {
	var err error
	// Based on code from github.com/docker/cli/context/docker/load.go
	// TODO (bfirsh): support reading from different SSH keys (with GCloud one by default)
	endpoint := dockerContext.Endpoint{
		EndpointMeta: dockerContext.EndpointMeta{
			Host: host,
		},
	}
	clientOpts, err := endpoint.ClientOpts()
	if err != nil {
		return nil, err
	}

	return client.NewClientWithOpts(clientOpts...)
}

// Based on containerAttach in github.com/docker/cli cli/command/container/run.go, but using logs instead of attach
func connectToLogs(ctx context.Context, dockerClient *client.Client, errChan *chan error, containerID string) (closeFunc, error) {
	response, err := dockerClient.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		return nil, err
	}

	ch := make(chan error, 1)
	*errChan = ch

	go func() {
		ch <- func() error {
			_, errCopy := stdcopy.StdCopy(os.Stdout, os.Stderr, response)
			return errCopy
		}()
	}()

	return response.Close, nil
}

// Based on waitExitOrRemoved in github.com/docker/cli cli/command/container/utils.go
func waitUntilExit(ctx context.Context, dockerClient *client.Client, containerID string) <-chan int {
	// TODO check for API version >=1.30

	resultChan, errChan := dockerClient.ContainerWait(ctx, containerID, container.WaitConditionNextExit)

	statusChan := make(chan int)
	go func() {
		select {
		case result := <-resultChan:
			if result.Error != nil {
				console.Error("Error waiting for container: %v", result.Error.Message)
				statusChan <- 125
			} else {
				statusChan <- int(result.StatusCode)
			}
		case err := <-errChan:
			console.Error("error waiting for container: %v", err)
			statusChan <- 125
		}
	}()

	return statusChan
}

func findSourceDir() (string, error) {
	if global.SourceDirectory == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return "", err
		}
		configPath, err := config.FindConfigPath(cwd)
		if err != nil {
			console.Debug("%s", err)
			return cwd, nil
		}
		return filepath.Dir(configPath), nil
	}
	return global.SourceDirectory, nil
}

func parseHost(hostWithUsernameAndPort string) (username string, host string, port int, err error) {
	re := regexp.MustCompile("^(?:([^@]+)@)?([^:]+)(?:([0-9]+))?$")
	matches := re.FindStringSubmatch(hostWithUsernameAndPort)
	if len(matches) == 0 {
		return "", "", 0, fmt.Errorf("Invalid host. The host must be in the format [username@]hostname[:port]")
	}
	username = matches[1]
	host = matches[2]
	port, err = strconv.Atoi(matches[3])
	if err != nil {
		return "", "", 0, err
	}
	return username, host, port, nil
}
