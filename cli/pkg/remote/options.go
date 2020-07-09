package remote

import (
	"fmt"
	"time"

	"replicate.ai/cli/pkg/files"
)

// TODO: password login

type Options struct {
	Host           string
	Port           int
	Username       string
	PrivateKeys    []string
	ConnectTimeout time.Duration
}

func (o *Options) SSHArgs() []string {
	args := []string{}

	if o.PrivateKeys != nil {
		for _, key := range o.PrivateKeys {
			args = append(args, "-i", key)
		}
	}
	for _, keyPath := range DefaultPrivateKeys() {
		absPath, err := files.ExpandUser(keyPath)
		if err != nil {
			panic("default private key has an error")
		}
		args = append(args, "-i", absPath)
	}

	if o.Port != 0 {
		args = append(args, "-p", fmt.Sprintf("%d", o.Port))
	}

	// TODO: host key verification
	args = append(args,
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "StrictHostKeyChecking=no",
	)

	// TODO: lower log level?
	args = append(args, "-o", "LogLevel=ERROR")

	return args
}
