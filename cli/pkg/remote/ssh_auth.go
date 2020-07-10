package remote

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"runtime"
	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/files"
)

func DefaultPrivateKeys() []string {
	return []string{
		"~/.ssh/id_rsa",
		"~/.ssh/id_dsa",
		"~/.ssh/google_compute_engine",
	}
}

func authMethodsFromOptions(options *Options) ([]ssh.AuthMethod, error) {
	// concatenate as many auth methods as we can
	authMethods := []ssh.AuthMethod{}

	if options.PrivateKeys != nil {
		for _, path := range options.PrivateKeys {
			pkAuth, err := privateKeyAuthMethod(path)
			if err != nil {
				return nil, err
			}
			authMethods = append(authMethods, pkAuth)
		}
	}

	for _, path := range DefaultPrivateKeys() {
		absPath, err := files.ExpandUser(path)
		if err != nil {
			panic("default private key has an error")
		}
		exists, err := files.FileExists(absPath)
		if err != nil {
			return nil, err
		}
		if exists {
			pkAuth, err := privateKeyAuthMethod(absPath)
			if err != nil {
				// HACK (bfirsh): ignore encrypted keys to fall back to SSH agent
				if strings.Contains(err.Error(), "cannot decode encrypted private keys") {
					continue
				}
				return nil, err
			}
			authMethods = append(authMethods, pkAuth)
		}
	}

	// always add ssh agent auth if we can, but allow it to
	// not be available
	agentAuth, err := sshAgentAuthMethod()
	if err != nil {
		console.Debug("SSH agent auth is not available: %s", err)
	} else {
		authMethods = append(authMethods, agentAuth)
	}

	return authMethods, nil
}

func privateKeyAuthMethod(privateKeyPath string) (ssh.AuthMethod, error) {
	contents, err := ioutil.ReadFile(privateKeyPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to read private key at %s, got error: %s", privateKeyPath, err)
	}
	signer, err := ssh.ParsePrivateKey(contents)
	if err != nil {
		return nil, fmt.Errorf("Failed to parse private key at %s, got error: %s", privateKeyPath, err)
	}
	return ssh.PublicKeys(signer), nil
}

func sshAgentAuthMethod() (ssh.AuthMethod, error) {
	switch runtime.GOOS {
	case "linux":
		return sshAgentUnixAuthMethod()
	case "windows":
		// TODO(andreas): something like https://github.com/sfreiberg/simplessh/blob/master/agent_windows.go
		return nil, fmt.Errorf("SSH agent connection is not available on Windows")
	case "darwin":
		return sshAgentUnixAuthMethod()
	default:
		return nil, fmt.Errorf("SSH agent connection is not available on the %s operating system", runtime.GOOS)
	}
}

func sshAgentUnixAuthMethod() (ssh.AuthMethod, error) {
	sshAgent, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
	if err != nil {
		return nil, fmt.Errorf("Failed to get SSH agent from SSH_AUTH_SOCK")
	}

	return ssh.PublicKeysCallback(agent.NewClient(sshAgent).Signers), nil
}
