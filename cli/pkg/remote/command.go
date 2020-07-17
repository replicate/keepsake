package remote

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/kballard/go-shellquote"
	"golang.org/x/crypto/ssh"

	"replicate.ai/cli/pkg/slices"
)

// if we passthrough the environment to the remote host, we must
// filter some variables that cause problems. TODO(andreas): this is
// brittle. maybe use a whitelist?
var envBlacklist = []string{"Apple_PubSub_Socket_Render", "COLUMNS", "COMMAND_MODE", "DISPLAY", "EDITOR", "HISTCONTROL", "HISTFILESIZE", "HISTSIZE", "HOME", "HOMEBREW_AUTO_UPDATE_SECS", "JAVA_HOME", "JICOFO_AUTH_PASSWORD", "JICOFO_COMPONENT_SECRET", "JVB_AUTH_PASSWORD", "LANG", "LC_ALL", "LC_CTYPE", "LOGNAME", "NODE_PATH", "OLDPWD", "PAGER", "PATH", "PS1", "PWD", "PYENV_SHELL", "PYENV_VIRTUALENV_INIT", "SDKMAN_PLATFORM", "SECURITYSESSIONID", "SHELL", "SHLVL", "SSH_AUTH_SOCK", "TERM", "TERMCAP", "TMPDIR", "USER", "XPC_FLAGS", "_", "__CF_USER_TEXT_ENCODING", "PYENV_ROOT", "PYENV_VERSION", "SDKMAN_CANDIDATES_API", "PYENV_DIR", "SDKMAN_VERSION", "PYENV_HOOK_PATH", "XPC_SERVICE_NAME", "SDKMAN_DIR", "SDKMAN_CANDIDATES_DIR", "PYTEST_CURRENT_TEST"}

type WrappedCmd struct {
	client  *Client
	cmd     *exec.Cmd
	session *ssh.Session
	env     map[string]string
}

// WrapCommand wraps an exec.Command with ssh execution,
// behind a similar API
func (c *Client) WrapCommand(cmd *exec.Cmd) *WrappedCmd {
	return &WrappedCmd{
		cmd:    cmd,
		client: c,
	}
}

// Command creates a new command similar to exec.Command,
// with ssh execution
func (c *Client) Command(name string, arg ...string) *WrappedCmd {
	return c.WrapCommand(exec.Command(name, arg...))
}

func (c *WrappedCmd) Output() ([]byte, error) {
	if err := c.newSession(); err != nil {
		return nil, err
	}
	cmdLine := c.getCommandLine()
	return c.session.Output(cmdLine)
}

func (c *WrappedCmd) CombinedOutput() ([]byte, error) {
	if err := c.newSession(); err != nil {
		return nil, err
	}
	cmdLine := c.getCommandLine()
	return c.session.CombinedOutput(cmdLine)
}

func (c *WrappedCmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

func (c *WrappedCmd) Start() error {
	if err := c.newSession(); err != nil {
		return err
	}
	cmdLine := c.getCommandLine()
	if err := c.session.Start(cmdLine); err != nil {
		c.session.Close()
		return err
	}

	return nil
}

func (c *WrappedCmd) Wait() error {
	defer c.session.Close()
	return c.session.Wait()
}

func (c *WrappedCmd) newSession() error {
	if c.session != nil {
		// panic since this is a programming error
		panic("Session is already started")
	}

	var err error
	c.session, err = c.client.sshClient.NewSession()
	if err != nil {
		return err
	}

	c.session.Stdin = c.cmd.Stdin
	c.session.Stdout = c.cmd.Stdout
	c.session.Stderr = c.cmd.Stderr

	c.env = FilterEnvMap(c.cmd.Env)
	return nil
}

func (c *WrappedCmd) getCommandLine() string {
	cmdLine := shellquote.Join(c.cmd.Args...)

	if c.cmd.Dir != "" {
		cmdLine = shellquote.Join("cd", c.cmd.Dir) + "; " + cmdLine
	}

	// session.setEnv doesn't work if AcceptEnv isn't set on
	// the SSH server, so we have to hack together an export
	// string instead.
	if len(c.env) > 0 {
		exports := []string{}
		for name, value := range c.env {
			exports = append(exports, fmt.Sprintf("%s=%s", name, shellquote.Join(value)))
		}
		cmdLine = fmt.Sprintf("export %s; %s", strings.Join(exports, " "), cmdLine)
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
