package docker

import (
	"io"
	"os"
	"os/exec"
	"strings"

	"replicate.ai/cli/pkg/console"
)

// Build a Docker image by calling `docker build` locally
//
// Log output is sent to stdout/err.
func Build(host string, path string, dockerfile string, name string) error {
	// TODO: This is local just to get this working. It probably wants to be remote so Docker doesn't
	// have to be installed locally, and so SSH keys can work

	args := []string{}
	if host != "" {
		args = append(args, "--host", host)
	}
	args = append(args,
		"build", ".",
		"--build-arg", "BUILDKIT_INLINE_CACHE=1",
		"--file", "-",
		"--tag", name,
	)

	cmd := exec.Command("docker", args...)
	cmd.Dir = path
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DOCKER_BUILDKIT=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	go func() {
		defer stdin.Close()
		io.WriteString(stdin, dockerfile)
	}()

	console.Debug("Running '%v'", strings.Join(cmd.Args, " "))

	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}
