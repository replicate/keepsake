package experiment

import (
	"encoding/json"
	"path"
	"time"

	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

// corresponds to DEFAULT_REFRESH_INTERVAL in heartbeat.py
var heartbeatRefreshInterval = 10 * time.Second

// the number of missed heartbeats we tolerate before declaring
// the experiment "stopped"
var heartbeatMissTolerance = 3

// Experiment represents a training run
type Experiment struct {
	ID      string                  `json:"id"`
	Created time.Time               `json:"created"`
	Params  map[string]*param.Value `json:"params"`
	Host    string                  `json:"host"`
	User    string                  `json:"user"`

	// properties that are not actually part of metadata json
	LastHeartbeat time.Time `json:"-"`
	Running       bool      `json:"-"`
}

// NewExperiment creates a commit, setting ID and Created
func NewExperiment(params map[string]*param.Value) *Experiment {
	return &Experiment{
		ID:      hash.Random(),
		Created: time.Now().UTC(),
		Params:  params,
	}
}

// Save experiment to storage
func (e *Experiment) Save(storage storage.Storage) error {
	data, err := json.MarshalIndent(e, "", " ")
	if err != nil {
		return err
	}
	return storage.Put(path.Join("experiments", e.ID, "replicate-metadata.json"), data)
}

func (e *Experiment) Heartbeat(storage storage.Storage, t time.Time) error {
	heartbeat := &Heartbeat{
		ExperimentID:  e.ID,
		LastHeartbeat: t,
	}
	data, err := json.MarshalIndent(heartbeat, "", " ")
	if err != nil {
		return err
	}
	return storage.Put(path.Join("experiments", e.ID, "replicate-heartbeat.json"), data)
}

func IsRunning(lastHeartbeat time.Time) bool {
	now := time.Now()
	lastTolerableHeartbeat := now.Add(-heartbeatRefreshInterval * time.Duration(heartbeatMissTolerance))
	return lastHeartbeat.After(lastTolerableHeartbeat)
}
