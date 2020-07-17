package docker

import (
	"io"
	"os"
	"os/exec"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/remote"
	"replicate.ai/cli/pkg/rsync"
)

// Build a Docker image by calling `docker build` locally or remotely over SSH
//
// Log output is sent to stdout/err.
func Build(remoteOptions *remote.Options, folder string, dockerfile string, name string, baseImage string, hasGPU bool) error {
	args := []string{
		"build", ".",
		"--build-arg", "BUILDKIT_INLINE_CACHE=1",
		"--build-arg", "BASE_IMAGE=" + baseImage,
		"--progress", "plain",
		"--file", "-",
		"--tag", name,
	}
	if hasGPU {
		args = append(args, "--build-arg", "HAS_GPU=1")
	}

	cmd := exec.Command("docker", args...)
	cmd.Env = remote.FilterEnvList(os.Environ())
	cmd.Env = append(cmd.Env, "DOCKER_BUILDKIT=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, dockerfile) //nolint
	}()

	console.Debug("Running '%s'", cmd.String())

	if remoteOptions == nil {
		cmd.Dir = folder
		if err := cmd.Start(); err != nil {
			return err
		}
		return cmd.Wait()
	}

	remoteTempDir, err := rsync.UploadToTempDir(folder, remoteOptions)
	if err != nil {
		return err
	}
	cmd.Dir = remoteTempDir
	client, err := remote.NewClient(remoteOptions)
	if err != nil {
		return err
	}
	remoteCmd := client.WrapCommand(cmd)
	if err := remoteCmd.Start(); err != nil {
		return err
	}
	return remoteCmd.Wait()
}
