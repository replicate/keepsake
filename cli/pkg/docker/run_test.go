package docker

import (
	"io/ioutil"
	"os/exec"
	"path"
	"testing"

	"github.com/kami-zh/go-capturer"
	"github.com/stretchr/testify/require"
)

func TestRunMounts(t *testing.T) {
	dockerClient, err := NewLocalClient()
	require.NoError(t, err)

	tmpdir, err := ioutil.TempDir("/tmp", "test-run")
	require.NoError(t, err)

	err = ioutil.WriteFile(path.Join(tmpdir, "hello.txt"), []byte("hello\n"), 0644)
	require.NoError(t, err)

	mounts := []Mount{{
		HostDir:      tmpdir,
		ContainerDir: "/mounted",
	}}

	// docker.Run doesn't pull
	require.NoError(t, exec.Command("docker", "pull", "alpine").Run())

	out := capturer.CaptureStdout(func() {
		err = Run(dockerClient, "alpine", []string{"cat", "/mounted/hello.txt"}, mounts, false, "user", "host", "s3://storage-bucket")
	})
	require.NoError(t, err)
	require.Equal(t, "hello\n", out)
}
