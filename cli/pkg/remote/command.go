package remote

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"

	"github.com/adjust/uniuri"
	"github.com/kballard/go-shellquote"

	"replicate.ai/cli/pkg/slices"
)

// if we passthrough the environment to the remote host, we must
// filter some variables that cause problems. TODO(andreas): this is
// brittle. maybe use a whitelist?
var envBlacklist = []string{"Apple_PubSub_Socket_Render", "COLUMNS", "COMMAND_MODE", "DISPLAY", "EDITOR", "GOOGLE_APPLICATION_CREDENTIALS", "HISTCONTROL", "HISTFILESIZE", "HISTSIZE", "HOME", "HOMEBREW_AUTO_UPDATE_SECS", "JAVA_HOME", "JICOFO_AUTH_PASSWORD", "JICOFO_COMPONENT_SECRET", "JVB_AUTH_PASSWORD", "LANG", "LC_ALL", "LC_CTYPE", "LOGNAME", "NODE_PATH", "OLDPWD", "PAGER", "PATH", "PS1", "PWD", "PYENV_SHELL", "PYENV_VIRTUALENV_INIT", "SDKMAN_PLATFORM", "SECURITYSESSIONID", "SHELL", "SHLVL", "SSH_AUTH_SOCK", "TERM", "TERMCAP", "TMPDIR", "USER", "XPC_FLAGS", "_", "__CF_USER_TEXT_ENCODING", "PYENV_ROOT", "PYENV_VERSION", "SDKMAN_CANDIDATES_API", "PYENV_DIR", "SDKMAN_VERSION", "PYENV_HOOK_PATH", "XPC_SERVICE_NAME", "SDKMAN_DIR", "SDKMAN_CANDIDATES_DIR", "PYTEST_CURRENT_TEST"}

// WrapCommand wraps an exec.Command with ssh execution,
// behind a similar API
func (c *Client) WrapCommand(cmd *exec.Cmd) *exec.Cmd {
	cmdLine := getCommandLine(cmd)
	// SendEnv doesn't work if AcceptEnv isn't set on
	// the SSH server, so we have to hack together an export
	// string instead.
	// TODO: is there a better way to do this?
	if len(cmd.Env) > 0 {
		exports := []string{}
		for _, env := range cmd.Env {
			parts := strings.SplitN(env, "=", 2)
			// TODO(andreas): check that parts is well-formed <name>=<value>
			name := parts[0]
			value := parts[1]
			exports = append(exports, fmt.Sprintf("%s=%s", name, shellquote.Join(value)))
		}
		cmdLine = fmt.Sprintf("export %s; %s", strings.Join(exports, " "), cmdLine)
	}

	args := c.options.SSHArgs()
	args = append(args, c.options.Host, cmdLine)
	wrapped := exec.Command("ssh", args...)

	wrapped.Stdin = cmd.Stdin
	wrapped.Stdout = cmd.Stdout
	wrapped.Stderr = cmd.Stderr
	// Environment passed to local `ssh` command, not to command running on remote machine
	wrapped.Env = os.Environ()
	return wrapped
}

// WrapCommandExternalEnv writes environment variables to a temporary
// on the remote host. Before running the remote command, the
// temporary file is sourced, then immediately deleted.
// FIXME (bfirsh): this is no longer used, so remove if we turn out not to use it
func (c *Client) WrapCommandSafeEnv(cmd *exec.Cmd) (*exec.Cmd, error) {
	remoteEnvPath := path.Join("/tmp", uniuri.New()+".sh")
	envFile, err := c.sftpClient.Create(remoteEnvPath)
	if err != nil {
		return nil, fmt.Errorf("Failed to create environment file on remote host: %w", err)
	}
	contents := strings.Join(FilterEnvList(cmd.Env), "\n")
	if _, err = envFile.Write([]byte(contents)); err != nil {
		return nil, fmt.Errorf("Failed to write environment file on remote host: %w", err)
	}
	args := c.options.SSHArgs()
	cmdLine := fmt.Sprintf("source %q && rm %q && %s", remoteEnvPath, remoteEnvPath, getCommandLine(cmd))
	args = append(args, c.options.Host, cmdLine)
	wrapped := exec.Command("ssh", args...)

	wrapped.Stdin = cmd.Stdin
	wrapped.Stdout = cmd.Stdout
	wrapped.Stderr = cmd.Stderr

	return wrapped, nil
}

// Command creates a new command similar to exec.Command,
// with ssh execution
func (c *Client) Command(name string, arg ...string) *exec.Cmd {
	return c.WrapCommand(exec.Command(name, arg...))
}

func getCommandLine(cmd *exec.Cmd) string {
	//cmdLine := shellquote.Join(cmd.Args...)
	cmdLine := strings.Join(cmd.Args, " ")

	if cmd.Dir != "" {
		cmdLine = shellquote.Join("cd", cmd.Dir) + "; " + cmdLine
	}

	return cmdLine
}

// FilterEnvMap returns a map of environment variables, with
// blacklisted keys filtered out
func FilterEnvMap(env []string) map[string]string {
	filtered := map[string]string{}
	for _, env := range env {
		parts := strings.SplitN(env, "=", 2)
		name := parts[0]
		if slices.ContainsString(envBlacklist, name) {
			continue
		}

		value := ""
		if len(parts) == 2 {
			value = parts[1]
		}
		filtered[name] = value
	}
	return filtered
}

// FilterEnvList returns a list of "="-separated environment
// key/values, with blacklisted keys filtered out
func FilterEnvList(env []string) []string {
	filtered := FilterEnvMap(env)
	list := []string{}
	for key, val := range filtered {
		list = append(list, key+"="+val)
	}
	return list
}
