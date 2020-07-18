package experiment

import (
	"encoding/json"
	"path"
	"time"

	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

// Experiment represents a training run
type Experiment struct {
	ID        string                  `json:"id"`
	Timestamp float64                 `json:"timestamp"`
	Params    map[string]*param.Value `json:"params"`
}

// NewExperiment creates a commit, setting ID and Timestamp
func NewExperiment(params map[string]*param.Value) *Experiment {
	return &Experiment{
		ID:        hash.Random(),
		Timestamp: float64(time.Now().Unix()),
		Params:    params,
	}
}

// Save experiment to storage
func (e *Experiment) Save(storage storage.Storage) error {
	data, err := json.Marshal(e)
	if err != nil {
		return err
	}
	return storage.Put(path.Join("experiments", e.ID, "replicate-metadata.json"), data)
}
