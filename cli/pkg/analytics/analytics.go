package analytics

import (
	"io/ioutil"

	golog "log"

	segment "gopkg.in/segmentio/analytics-go.v3"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/global"
)

// Client wraps up the analytics clients we use
type Client struct {
	email         string
	segmentClient segment.Client
}

// Config configures the analytics client
type Config struct {
	Email      string
	SegmentKey string
}

// NewClient makes a new Client struct and instantiates the various other clients
func NewClient(config *Config) (client *Client) {
	client = &Client{email: config.Email}
	segmentClient, err := segment.NewWithConfig(config.SegmentKey, segment.Config{
		// Flush immediately
		BatchSize: 1,
		// TODO: hook up to global.Verbose
		Verbose: true,
		Logger:  segment.StdLogger(golog.New(ioutil.Discard, "segment", golog.LstdFlags)),
	})
	// Ignore broken Segment -- it isn't essential
	if err != nil {
		console.Warn("Error settings up Segment client: %v", err)
	} else {
		client.segmentClient = segmentClient
	}
	return client
}

// Track an event
func (c *Client) Track(event string, properties map[string]string) {
	if c.segmentClient == nil {
		return
	}
	segmentProperties := segment.NewProperties()
	segmentProperties.Set("version", global.Version)

	for k, v := range properties {
		segmentProperties.Set(k, v)
	}
	c.segmentClient.Enqueue(segment.Track{
		UserId:     c.email,
		Event:      event,
		Properties: segmentProperties,
	})
}

// Close and flush any clients
//
// FIXME (bfirsh): short timeout
func (c *Client) Close() error {
	if c.segmentClient != nil {
		return c.segmentClient.Close()
	}
	return nil
}
