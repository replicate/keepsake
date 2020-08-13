package docker

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/files"
	"replicate.ai/cli/pkg/remote"
)

// Build a Docker image by calling `docker build` locally or remotely over SSH
//
// Log output is sent to stdout/err.
func Build(remoteOptions *remote.Options, folder string, dockerfile string, name string, baseImage string, hasGPU bool) error {

	args := []string{
		"build", ".",
		"--build-arg", "BUILDKIT_INLINE_CACHE=1",
		"--build-arg", "BASE_IMAGE=" + baseImage,
		"--tag", name,
	}
	// TODO(andreas): detect if terminal supports cursor movement
	if !console.IsTTY() {
		args = append(args, "--progress", "plain")
	}
	if hasGPU {
		args = append(args, "--build-arg", "HAS_GPU=1")
	}

	cmd := exec.Command("docker", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// OpenSSH must have stdin connected, because that's how it determines if there's a TTY attached.
	// If stdin is not a TTY, it disables TTY on stdout/err, meaning you can't pipe things to OpenSSH and
	// get TTY output (grr!).
	cmd.Stdin = os.Stdin

	console.Debug("Running '%s'", cmd.String())

	// Local
	if remoteOptions == nil {
		// This is tempdir not tempfile because github actions doesn't work with tempfiles.
		tmpDir, err := files.TempDir("dockerfile")
		if err != nil {
			return fmt.Errorf("Failed to create Dockerfile: %w", err)
		}
		dockerfilePath := path.Join(tmpDir, "Dockerfile")
		if err = ioutil.WriteFile(dockerfilePath, []byte(dockerfile), 0644); err != nil {
			return fmt.Errorf("Failed to write Dockerfile: %w", err)
		}
		cmd.Args = append(cmd.Args, "--file", dockerfilePath)

		// Pass through entire env, because "docker build" needs some basic envvars set
		cmd.Env = os.Environ()
		cmd.Env = append(cmd.Env, "DOCKER_BUILDKIT=1")
		cmd.Dir = folder
		if err := cmd.Start(); err != nil {
			return err
		}
		return cmd.Wait()
	}

	// Remote, via SSH

	// SSH runs this command inside a shell so it will have basic envvars set, so just
	// set the additional envvars we want
	cmd.Env = []string{"DOCKER_BUILDKIT=1"}

	remoteTempDir, err := remote.UploadToTempDir(folder, remoteOptions)
	if err != nil {
		return err
	}
	cmd.Dir = remoteTempDir
	client, err := remote.NewClient(remoteOptions)
	if err != nil {
		return err
	}

	// Write Dockerfile
	dockerfileFile, err := client.SFTP().Create(path.Join(remoteTempDir, "Dockerfile"))
	if err != nil {
		return fmt.Errorf("Failed to create Dockerfile: %w", err)
	}
	if _, err = dockerfileFile.Write([]byte(dockerfile)); err != nil {
		return fmt.Errorf("Failed to write Dockerfile: %w", err)
	}
	if err := dockerfileFile.Close(); err != nil {
		return fmt.Errorf("Failed to write Dockerfile: %w", err)
	}

	// Docker build just writes to stderr -- this how it checks if it should show progress
	remoteCmd := client.WrapCommand(cmd)
	if err := remoteCmd.Start(); err != nil {
		return err
	}
	return remoteCmd.Wait()
}
