package remote

import (
	"bytes"
	"io"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

// TODO: test ssh agent auth

func TestExecOutput(t *testing.T) {
	mockRemote, err := NewMockRemote()
	require.NoError(t, err)
	defer mockRemote.Kill()

	client, err := NewClient(&Options{
		Host:        "localhost",
		Port:        mockRemote.Port,
		Username:    "root",
		PrivateKeys: []string{mockRemote.PrivateKeyPath},
	})
	require.NoError(t, err)

	remoteCmd := client.Command("whoami")
	output, err := remoteCmd.Output()
	require.NoError(t, err)

	require.Equal(t, "root\n", string(output))
}

func TestExecInput(t *testing.T) {
	mockRemote, err := NewMockRemote()
	require.NoError(t, err)
	defer mockRemote.Kill()

	client, err := NewClient(&Options{
		Host:        "localhost",
		Port:        mockRemote.Port,
		Username:    "root",
		PrivateKeys: []string{mockRemote.PrivateKeyPath},
	})
	require.NoError(t, err)

	cmd := exec.Command("cat", "-")
	stdin, err := cmd.StdinPipe()
	require.NoError(t, err)

	output := bytes.Buffer{}
	cmd.Stdout = &output

	remoteCmd := client.WrapCommand(cmd)
	err = remoteCmd.Start()
	require.NoError(t, err)

	_, err = io.WriteString(stdin, "hello world\n")
	require.NoError(t, err)
	stdin.Close()

	remoteCmd.Wait()

	require.Equal(t, "hello world\n", output.String())
}
