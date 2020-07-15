package remote

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/netutils"
)

type MockRemote struct {
	Port           int
	ContainerName  string
	PrivateKeyPath string
}

// NewMockRemote starts a docker container with an SSH server,
// with a private key in mockRemote.PrivateKeyPath.
// call mockRemote.Kill() to stop and delete the container
// and remove the private key.
func NewMockRemote() (*MockRemote, error) {
	m := new(MockRemote)

	port, err := netutils.NextFreePort(rand.Intn(20000) + 2000)
	if err != nil {
		return nil, err
	}
	m.Port = port

	m.ContainerName = hash.Random()
	dockerCmd := exec.Command("docker", "run", "-v", "/var/run/docker.sock:/var/run/docker.sock", "--rm", "--name", m.ContainerName, "-i", fmt.Sprintf("-p%d:22", port), "docker.io/andreasjansson/sshd:latest")
	dockerCmd.Stderr = os.Stderr
	dockerCmd.Stdout = os.Stderr
	if err = dockerCmd.Start(); err != nil {
		return nil, err
	}

	err = netutils.WaitForPort(port, 60*time.Second)
	if err != nil {
		m.Kill() //nolint
		return nil, fmt.Errorf("Failed to connect to sshd server before timeout")
	}

	// drop public key
	cmd := exec.Command("docker", "exec", "-i", m.ContainerName, "bash", "-c", "mkdir -p /root/.ssh && cat - > /root/.ssh/authorized_keys && chmod 400 /root/.ssh/authorized_keys")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		m.Kill() //nolint
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		m.Kill() //nolint
		return nil, err
	}
	if _, err := stdin.Write([]byte(publicKey)); err != nil {
		m.Kill() //nolint
		return nil, err
	}
	if err := stdin.Close(); err != nil {
		m.Kill() //nolint
		return nil, err
	}
	if err := cmd.Wait(); err != nil {
		m.Kill() //nolint
		return nil, err.(*exec.ExitError)
	}

	privateKeyFile, err := ioutil.TempFile("/tmp", "replicate-test-private-key-*")
	if err != nil {
		m.Kill() //nolint
		return nil, err
	}
	if _, err := privateKeyFile.WriteString(privateKey); err != nil {
		m.Kill() //nolint
		return nil, err
	}
	if err := privateKeyFile.Close(); err != nil {
		m.Kill() //nolint
		return nil, err
	}
	m.PrivateKeyPath = privateKeyFile.Name()

	return m, nil
}

func (m *MockRemote) Kill() error {
	os.Remove(m.PrivateKeyPath)

	err := exec.Command("docker", "kill", m.ContainerName).Run()
	if err != nil {
		// log an error since Kill() will be called defered in tests
		console.Error("Failed to kill %s, got error: %s", m.ContainerName, err)
		return err
	}
	return nil
}

const privateKey = `-----BEGIN RSA PRIVATE KEY-----
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
-----END RSA PRIVATE KEY-----
`

const publicKey = `ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQDLK9w4BtrqVxT4hl5guUzFM8jCMCxOaGKghr/nkGh3dbhabbHNjxNb0lCRJQMVwGSp/Ljb52oQZv9QqpAWBMVDOyuZ/7Z/ftU+TSTZVENiCGKG2KiScC7GDlINq0mdbh/cq6Ly4uzxrMA+mNo0UmYQT43A+tcKzZjobitLNOwhPa+5UO0m0I5XY5RN7IDexXDAVQ5Z6rPMg8cB3yE9j/kbwDYDkeUAxpPTK8Pcm7ELdMeTR3Bfjqji4vgsXirSHQKZir2qRHkDpeTywELLEVRrJFAJV9VkyMe6Q4v6X+do0iZyaknH6cUzkzBMHGXSTYkVcwQl67EWVmmWApnA0eqs0iITf7nLM3kFcdmVHK0R4AO0qm8wiDC6d3Cs/bJe8WA1rkVgrVGV4HA3nt6quWUIXLe96ukjpsaPNL6r4pvtFlET2dm9Sw7MjKK/0YzkbPdJgdmL1k3PwrsdK8YoXf/bViQ2KsN6hKqPG5zL7286aHxjlTiRMCfLiTYiBxZHpBdG6KZL45XZpF7MvbSfUHUDTAWr/Wb7O2vbof21SZ3sgGrYL8KjKAdUD6FPN/RTeD78Clc5Jw3KrCT39PXrtN57vJ7gx+u6MfDV6buumoLfAQZZWTRXvulLMedUr4e00trD8dwwQWcGA4GtGkv7W8p6HI8UTnd3GcmGvDurKNptLQ== andreas@Andreass-MacBook-Pro.local`
