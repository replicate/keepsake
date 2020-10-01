package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kballard/go-shellquote"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/remote"
	"replicate.ai/cli/pkg/storage"
)

type closeFunc func() error

type Mount struct {
	HostDir      string
	ContainerDir string
}

// Run runs a Docker container from imageName with cmd
// TODO: this could do with a configuration struct
// TODO(andreas): this function is getting really unwieldy and has lots of responsibilities, let's refactor
func Run(dockerClient *client.Client, imageName string, cmd []string, mounts []Mount, hasGPU bool, user string, host string, storageURL string, env []string) error {
	// use same name for both container and image
	containerName := imageName

	osEnv := remote.FilterEnvList(os.Environ())
	env = append(env, osEnv...)
	env = append(env, "PYTHONUNBUFFERED=1")

	// These environment variables are used by the
	// python library to save experiment metadata
	env = append(env, "REPLICATE_INTERNAL_USER="+user)
	env = append(env, "REPLICATE_INTERNAL_HOST="+host)
	env = append(env, "REPLICATE_INTERNAL_COMMAND="+shellquote.Join(cmd...))

	// Options for creating container
	config := &container.Config{
		Image: imageName,
		Cmd:   cmd,
		Env:   env,
	}
	// Options for starting container (port bindings, volume bindings, etc)
	hostConfig := &container.HostConfig{
		AutoRemove: true,
		Mounts:     []mount.Mount{},
	}
	if hasGPU {
		hostConfig.Runtime = "nvidia"
	}

	for _, mnt := range mounts {
		hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
			Source: mnt.HostDir,
			Target: mnt.ContainerDir,

			// TODO(andreas): are there cases where you might want to write to a directory on the host?
			ReadOnly: true,

			// TODO(andreas): is this the best mount type?
			Type: mount.TypeBind,
		})
	}

	// if storage is disk storage, it doesn't make sense to write
	// to the storage path inside the container since it's not
	// available outside the container. to fix this we rewrite the
	// storage path inside the container to a temporary directory,
	// and mount that to the actual storage path on the host.
	scheme, _, storagePath, err := storage.SplitURL(storageURL)
	if err != nil {
		return err
	}
	if scheme == storage.SchemeDisk {
		storagePath, err = filepath.Abs(storagePath)
		if err != nil {
			return fmt.Errorf("Failed to determine absolute directory of %s: %w", storagePath, err)
		}
		inContainerStorage := "/replicate-storage"
		config.Env = append(config.Env, "REPLICATE_STORAGE="+inContainerStorage)
		hostConfig.Mounts = append(hostConfig.Mounts, mount.Mount{
			Source:   storagePath,
			Target:   inContainerStorage,
			ReadOnly: false,
			Type:     mount.TypeBind,
		})
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
