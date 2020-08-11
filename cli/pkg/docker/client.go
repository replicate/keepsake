package docker

import (
	"net/http"
	"os"

	"github.com/docker/cli/cli/connhelper"
	dockerContext "github.com/docker/cli/cli/context/docker"
	"github.com/docker/docker/client"

	"replicate.ai/cli/pkg/remote"
)

// NewLocalClient creates a Docker client that connects to the default local Docker daemon
//
// The host is not yet configurable. It just lets Docker does whatever it does by default.
func NewLocalClient() (*client.Client, error) {
	endpoint := dockerContext.Endpoint{}
	clientOpts, err := endpoint.ClientOpts()
	if err != nil {
		return nil, err
	}
	return client.NewClientWithOpts(clientOpts...)
}

// NewRemoteClient creates a Docker client that connects via SSH
func NewRemoteClient(options *remote.Options) (*client.Client, error) {
	// Based on https://github.com/docker/cli/blob/0d26302d8a82030b71e34ae4c16e1168c24f866b/cli/context/docker/load.go#L93-L134
	// Following the code path where host is prefixed with "ssh://" and a connhelper is made

	args := options.SSHArgs()
	args = append(args, "--", options.Host)
	args = append(args, "docker", "system", "dial-stdio")

	helper, err := connhelper.GetCommandConnectionHelper("ssh", args...)
	if err != nil {
		return nil, err
	}

	var opts []client.Opt
	httpClient := &http.Client{
		// No tls
		// No proxy
		Transport: &http.Transport{
			DialContext: helper.Dialer,
		},
	}
	opts = append(opts,
		client.WithHTTPClient(httpClient),
		client.WithHost(helper.Host),
		client.WithDialContext(helper.Dialer),
	)
	version := os.Getenv("DOCKER_API_VERSION")
	if version != "" {
		opts = append(opts, client.WithVersion(version))
	} else {
		opts = append(opts, client.WithAPIVersionNegotiation())
	}

	return client.NewClientWithOpts(opts...)
}
