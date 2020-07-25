package experiment

import (
	"encoding/json"
	"fmt"
	"time"

	"replicate.ai/cli/pkg/storage"
)

type Heartbeat struct {
	ExperimentID  string    `json:"experiment_id"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
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
