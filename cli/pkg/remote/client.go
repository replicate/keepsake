package remote

import (
	"fmt"
	"net"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
)

type Client struct {
	options    *Options
	tcpConn    net.Conn
	sshConn    ssh.Conn
	sshClient  *ssh.Client
	sftpClient *sftp.Client
	session    ssh.Session
}

// TODO: intern connections keyed on options

// NewClient creates a new SSH client
func NewClient(options *Options) (*Client, error) {
	c := &Client{
		options: options,
	}

	port := 22
	if options.Port != 0 {
		port = options.Port
	}
	hostWithPort := net.JoinHostPort(options.Host, fmt.Sprintf("%d", port))

	tcpConn, err := net.DialTimeout("tcp", hostWithPort, options.ConnectTimeout)
	if err != nil {
		return nil, fmt.Errorf("Failed to connect to %s, got error: %s", hostWithPort, err)
	}
	c.tcpConn = tcpConn

	authMethods, err := authMethodsFromOptions(options)
	if err != nil {
		return nil, err
	}

	config := &ssh.ClientConfig{
		User:            options.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: handle host keys securely
	}
	sshConn, chans, reqs, err := ssh.NewClientConn(c.tcpConn, hostWithPort, config)
	if err != nil {
		return nil, fmt.Errorf("Failed to open SSH connection to %s, got error: %s", hostWithPort, err)
	}
	c.sshConn = sshConn
	c.sshClient = ssh.NewClient(sshConn, chans, reqs)

	c.sftpClient, err = sftp.NewClient(c.sshClient)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// SFTP returns a https://godoc.org/github.com/pkg/sftp#Client,
// capable of issuing filesystem commands remotely.
// For example: client.SFTP().Mkdir("foo")
func (c *Client) SFTP() *sftp.Client {
	return c.sftpClient
}

func (c *Client) WriteFile(data []byte, path string) error {
	remoteFile, err := c.sftpClient.Create(path)
	if err != nil {
		return fmt.Errorf("Failed to create remote file at %s, got error: %s", path, err)
	}
	if _, err := remoteFile.Write(data); err != nil {
		return fmt.Errorf("Failed to write remote file at %s, got error: %s", path, err)
	}
	if err := remoteFile.Close(); err != nil {
		return fmt.Errorf("Failed to close remote file at %s, got error: %s", path, err)
	}

	return nil
}
