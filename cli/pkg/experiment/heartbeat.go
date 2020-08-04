package experiment

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/storage"
)

// corresponds to DEFAULT_REFRESH_INTERVAL in heartbeat.py
var heartbeatRefreshInterval = 10 * time.Second

// the number of missed heartbeats we tolerate before declaring
// the experiment "stopped"
var heartbeatMissTolerance = 3

type Heartbeat struct {
	ExperimentID  string    `json:"experiment_id"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}

func CreateHeartbeat(storage storage.Storage, experimentID string, t time.Time) error {
	heartbeat := &Heartbeat{
		ExperimentID:  experimentID,
		LastHeartbeat: t,
	}
	data, err := json.MarshalIndent(heartbeat, "", " ")
	if err != nil {
		return err
	}
	return storage.Put(path.Join("metadata", "heartbeats", experimentID+".json"), data)
}

func listHeartbeats(store storage.Storage) ([]*Heartbeat, error) {
	paths, err := store.List("metadata/heartbeats/")
	if err != nil {
		return nil, err
	}
	heartbeats := []*Heartbeat{}
	for _, p := range paths {
		exp, err := loadHeartbeatFromPath(store, p)
		if err == nil {
			heartbeats = append(heartbeats, exp)
		} else {
			// TODO: should this just be ignored? can this be recovered from?
			console.Warn("Failed to load metadata from %q: %s", p, err)
		}
	}
	return heartbeats, nil
}

func (h *Heartbeat) IsRunning() bool {
	now := time.Now().UTC()
	lastTolerableHeartbeat := now.Add(-heartbeatRefreshInterval * time.Duration(heartbeatMissTolerance))
	return h.LastHeartbeat.After(lastTolerableHeartbeat)
}

func loadHeartbeatFromPath(store storage.Storage, path string) (*Heartbeat, error) {
	contents, err := store.Get(path)
	if err != nil {
		return nil, err
	}
	hb := new(Heartbeat)
	if err := json.Unmarshal(contents, hb); err != nil {
		return nil, fmt.Errorf("Parse error: %s", err)
	}
	return hb, nil
}
