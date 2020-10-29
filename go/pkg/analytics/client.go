package analytics

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	segment "github.com/segmentio/analytics-go"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/files"
)

// Event used for repository on disk.
type Event struct {
	Event      string                 `json:"event"`
	Properties map[string]interface{} `json:"properties"`
	Timestamp  time.Time              `json:"timestamp"`
}

// Config configures the analytics client
type Config struct {
	SegmentKey  string
	AnonymousID string
	Dir         string // Dir for storing state
}

// Client provides a thin layer over Segment's client for tracking CLI metrics.
//
// Events are buffered on disk to improve user experience, only periodically flushing to
// the Segemnt API.
//
// You should call Flush() when desired in order to flush to Segment. You may choose
// to do this after a certain number of events (see Size()) have been buffered,
// or after a given duration (see LastFlush()).
//
// Based on https://github.com/tj/go-cli-analytics
type Client struct {
	*Config
	eventsFile *os.File
	events     *json.Encoder
}

func NewClient(config *Config) (*Client, error) {
	a := &Client{Config: config}
	err := a.init()
	return a, err
}

func (a *Client) init() error {
	if err := os.Mkdir(a.Dir, 0755); err != nil && !os.IsExist(err) {
		return fmt.Errorf("error making %s: %w", a.Dir, err)
	}

	path := filepath.Join(a.Dir, "events")

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0666)
	if err != nil {
		return fmt.Errorf("analytics: error opening events: %w", err)
	}
	a.eventsFile = f
	a.events = json.NewEncoder(f)
	return nil
}

// Events reads the events from disk.
func (a *Client) Events() (v []*Event, err error) {
	f, err := os.Open(filepath.Join(a.Dir, "events"))
	if err != nil {
		return nil, errors.Wrap(err, "opening")
	}

	dec := json.NewDecoder(f)

	for {
		var e Event
		err := dec.Decode(&e)

		if err == io.EOF {
			break
		}

		if err != nil {
			// FIXME (bfirsh): Theoretically O_APPEND writes are atomic, but sometimes they aren't if they're too large
			// If a write collided, then this might cause decoding errors, then the events file is forever broken and
			// analytics events will never be sent again. This should probably delete the events buffer to fix the broken
			// write.
			return nil, errors.Wrap(err, "decoding")
		}

		v = append(v, &e)
	}

	return v, nil
}

// Size returns the number of events.
func (a *Client) Size() (int, error) {
	events, err := a.Events()
	if err != nil {
		return 0, errors.Wrap(err, "reading events")
	}

	return len(events), nil
}

// Touch ~/<dir>/last_flush.
func (a *Client) Touch() error {
	path := filepath.Join(a.Dir, "last_flush")
	return ioutil.WriteFile(path, []byte(""), 0755)
}

// LastFlush returns the last flush time.
func (a *Client) LastFlush() (time.Time, error) {
	info, err := os.Stat(filepath.Join(a.Dir, "last_flush"))
	if err != nil {
		return time.Unix(0, 0), err
	}

	return info.ModTime(), nil
}

// LastFlushDuration returns the last flush time delta.
func (a *Client) LastFlushDuration() (time.Duration, error) {
	lastFlush, err := a.LastFlush()
	if err != nil {
		return 0, nil
	}

	return time.Since(lastFlush), nil
}

// Track event `name` with optional `props`.
func (a *Client) Track(name string, props map[string]interface{}) error {
	if a.events == nil {
		return nil
	}

	return a.events.Encode(&Event{
		Event:      name,
		Properties: props,
		Timestamp:  time.Now(),
	})
}

// ConditionalFlush flushes if event count is above `aboveSize`, or age is `aboveDuration`,
// otherwise Close() is called and the underlying file(s) are closed.
func (a *Client) ConditionalFlush(aboveSize int, aboveDuration time.Duration) error {
	lastFlushExists, err := files.FileExists(filepath.Join(a.Dir, "last_flush"))
	if err != nil {
		return err
	}
	// Flush on first flush
	// This causes analytics events to be sent on the second run of the command, after the
	// user has had a chance to disable them.
	// The analytics client is not instantiated at all on the first run -- see track.go.
	if !lastFlushExists {
		console.Debug("analytics: flushing on first flush")
		return a.Flush()
	}

	age, err := a.LastFlushDuration()
	if err != nil {
		return err
	}

	size, err := a.Size()
	if err != nil {
		return err
	}

	switch {
	case size >= aboveSize:
		console.Debug("analytics: flushing due to size")
		return a.Flush()
	case age >= aboveDuration:
		console.Debug("analytics: flushing due to duration")
		return a.Flush()
	default:
		return a.Close()
	}
}

// Flush the events to Segment, removing them from disk.
// FIXME (bfirsh): two clients could potentially flush at the same time. Maybe this needs a lock, or something.
func (a *Client) Flush() error {
	if err := a.Close(); err != nil {
		return errors.Wrap(err, "closing")
	}

	events, err := a.Events()
	if err != nil {
		return errors.Wrap(err, "reading events")
	}

	client := segment.New(a.SegmentKey)

	for _, event := range events {
		err := client.Enqueue(segment.Track{
			Event:       event.Event,
			AnonymousId: a.AnonymousID,
			Properties:  event.Properties,
			Timestamp:   event.Timestamp,
		})
		if err != nil {
			return err
		}
	}

	if err := client.Close(); err != nil {
		return errors.Wrap(err, "closing client")
	}

	if err := a.Touch(); err != nil {
		return errors.Wrap(err, "touching")
	}

	return os.Remove(filepath.Join(a.Dir, "events"))
}

// Close the underlying file descriptor(s).
func (a *Client) Close() error {
	return a.eventsFile.Close()
}
