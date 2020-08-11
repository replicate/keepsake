package remote

import (
	"bytes"
	"io"
	"os/exec"
	"testing"

	"github.com/adjust/uniuri"
	"github.com/stretchr/testify/require"
)

// TODO: test ssh agent auth

func TestCommandOutput(t *testing.T) {
	mockRemote, err := NewMockRemote()
	require.NoError(t, err)
	defer mockRemote.Kill() //nolint

	client, err := NewClient(&Options{
		Host:        "localhost",
		Port:        mockRemote.Port,
		Username:    "root",
		PrivateKeys: []string{mockRemote.PrivateKeyPath},
	})
	require.NoError(t, err)

	output, err := client.Command("whoami").Output()
	require.NoError(t, err)

	require.Equal(t, "root\n", string(output))
}

func TestCommandInput(t *testing.T) {
	mockRemote, err := NewMockRemote()
	require.NoError(t, err)
	defer mockRemote.Kill() //nolint

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

	err = remoteCmd.Wait()
	require.NoError(t, err)

	require.Equal(t, "hello world\n", output.String())
}

func TestWrapCommandSafeEnv(t *testing.T) {
	mockRemote, err := NewMockRemote()
	require.NoError(t, err)
	defer mockRemote.Kill() //nolint

	client, err := NewClient(&Options{
		Host:        "localhost",
		Port:        mockRemote.Port,
		Username:    "root",
		PrivateKeys: []string{mockRemote.PrivateKeyPath},
	})
	require.NoError(t, err)

	myVar := uniuri.New()
	cmd := exec.Command("echo", "$MY_VAR")
	cmd.Env = append(cmd.Env, "MY_VAR="+myVar)
	wrapped, err := client.WrapCommandSafeEnv(cmd)
	require.NoError(t, err)
	output, err := wrapped.Output()
	require.NoError(t, err)

	require.Equal(t, myVar+"\n", string(output))

}

func TestWrapCommandUnsafeEnv(t *testing.T) {
	mockRemote, err := NewMockRemote()
	require.NoError(t, err)
	defer mockRemote.Kill() //nolint

	client, err := NewClient(&Options{
		Host:        "localhost",
		Port:        mockRemote.Port,
		Username:    "root",
		PrivateKeys: []string{mockRemote.PrivateKeyPath},
	})
	require.NoError(t, err)

	myVar := uniuri.New()
	cmd := exec.Command("echo", "$MY_VAR")
	cmd.Env = append(cmd.Env, "MY_VAR="+myVar)
	wrapped := client.WrapCommand(cmd)
	output, err := wrapped.Output()
	require.NoError(t, err)

	require.Equal(t, myVar+"\n", string(output))

}
