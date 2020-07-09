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

	text := uniuri.New()
	require.NoError(t, ioutil.WriteFile(path.Join(localDir, "foo.txt"), []byte(text), 0644))

	dockerfile := `
FROM alpine
COPY foo.txt /foo.txt
CMD cat /foo.txt
`
	name := "replicate-" + strings.ToLower(uniuri.NewLen(10))
	err = Build(nil, localDir, dockerfile, name)
	require.NoError(t, err)

	defer func() {
		if out, err := exec.Command("docker", "rmi", name).Output(); err != nil {
			console.Warn("Failed to remove docker image %s, got error: %s", name, out)
		}
	}()

	out, err := exec.Command("docker", "run", "-i", "--rm", name).CombinedOutput()
	require.NoError(t, err)
	require.Equal(t, text, string(out))
}

func TestBuildRemote(t *testing.T) {
	mockRemote, err := remote.NewMockRemote()
	require.NoError(t, err)
	defer mockRemote.Kill()

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
FROM alpine
COPY foo.txt /foo.txt
CMD cat /foo.txt
`
	client, err := remote.NewClient(options)
	require.NoError(t, err)

	name := "replicate-" + strings.ToLower(uniuri.NewLen(10))
	err = Build(options, localDir, dockerfile, name)
	require.NoError(t, err)

	defer func() {
		if out, err := client.Command("docker", "rmi", name).Output(); err != nil {
			console.Warn("Failed to remove docker image %s, got error: %s", name, out)
		}
	}()

	out, err := client.Command("docker", "run", "-i", "--rm", name).CombinedOutput()
	require.NoError(t, err)
	require.Equal(t, text, string(out))
}
