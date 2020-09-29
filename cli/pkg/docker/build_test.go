package docker

import (
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"

	"github.com/adjust/uniuri"
	"github.com/stretchr/testify/require"

	"os/exec"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/remote"
)

func TestBuildLocal(t *testing.T) {
	localDir, err := ioutil.TempDir("/tmp", "replicate-test-")
	require.NoError(t, err)
	defer os.RemoveAll(localDir)

	dockerfile := `
ARG BASE_IMAGE
FROM $BASE_IMAGE
ARG HAS_GPU
ENV HAS_GPU=$HAS_GPU
CMD echo $HAS_GPU
`
	name := "replicate-" + strings.ToLower(uniuri.NewLen(10))
	err = Build(nil, localDir, dockerfile, name, "alpine", true)
	require.NoError(t, err)

	defer func() {
		if out, err := exec.Command("docker", "rmi", name).Output(); err != nil {
			console.Warn("Failed to remove docker image %s: %s", name, out)
		}
	}()

	out, err := exec.Command("docker", "run", "-i", "--rm", name).CombinedOutput()
	require.NoError(t, err, string(out))
	require.Equal(t, "1\n", string(out))
}

func TestBuildRemote(t *testing.T) {
	mockRemote, err := remote.NewMockRemote()
	require.NoError(t, err)
	defer mockRemote.Kill() //nolint

	options := &remote.Options{
		Host:        "localhost",
		Port:        mockRemote.Port,
		Username:    "root",
		PrivateKeys: []string{mockRemote.PrivateKeyPath},
	}

	localDir, err := ioutil.TempDir("/tmp", "replicate-test-")
	require.NoError(t, err)
	defer os.RemoveAll(localDir)

	text := uniuri.New()
	require.NoError(t, ioutil.WriteFile(path.Join(localDir, "foo.txt"), []byte(text), 0644))

	dockerfile := `
ARG BASE_IMAGE
FROM $BASE_IMAGE
COPY foo.txt /foo.txt
CMD cat /foo.txt
`
	client, err := remote.NewClient(options)
	require.NoError(t, err)

	name := "replicate-" + strings.ToLower(uniuri.NewLen(10))
	err = Build(options, localDir, dockerfile, name, "alpine", false)
	require.NoError(t, err)

	defer func() {
		cmd := client.Command("docker", "rmi", name)
		require.NoError(t, err)
		if out, err := cmd.Output(); err != nil {
			console.Warn("Failed to remove docker image %s: %s", name, out)
		}
	}()

	cmd := client.Command("docker", "run", "-i", "--rm", name)
	require.NoError(t, err)
	out, err := cmd.Output()
	require.NoError(t, err)
	require.Equal(t, text, string(out))
}
