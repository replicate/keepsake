package experiment

import (
	"encoding/json"
	"fmt"
	"path"
	"time"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

// Experiment represents a training run
type Experiment struct {
	ID      string                  `json:"id"`
	Created time.Time               `json:"created"`
	Params  map[string]*param.Value `json:"params"`
	Host    string                  `json:"host"`
	User    string                  `json:"user"`
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
	return storage.Put(path.Join("metadata", "experiments", e.ID+".json"), data)
}

func List(store storage.Storage) ([]*Experiment, error) {
	paths, err := store.List("metadata/experiments/")
	if err != nil {
		return nil, err
	}
	experiments := []*Experiment{}
	for _, p := range paths {
		exp, err := loadExperimentFromPath(store, p)
		if err == nil {
			experiments = append(experiments, exp)
		} else {
			console.Warn("Failed to load metadata from %q: %s", p, err)
		}
	}
	return experiments, nil
}

func loadExperimentFromPath(store storage.Storage, path string) (*Experiment, error) {
	contents, err := store.Get(path)
	if err != nil {
		return nil, err
	}
	exp := new(Experiment)
	if err := json.Unmarshal(contents, exp); err != nil {
		return nil, fmt.Errorf("Parse error: %s", err)
	}
	return exp, nil
}
