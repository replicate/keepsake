package experiment

import (
	"encoding/json"
	"fmt"
	"path"
	"path/filepath"
	"time"

	"replicate.ai/cli/pkg/commit"
	"replicate.ai/cli/pkg/console"
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
	Commits []*commit.Commit        `json:"commits"`

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

func (e *Experiment) Commit(storage storage.Storage, metrics map[string]*param.Value, workingDir string) (*commit.Commit, error) {
	com := commit.NewCommit(e.ID, map[string]*param.Value{
		"accuracy": param.Float(0.987),
	})
	if err := com.Save(storage, workingDir); err != nil {
		return com, fmt.Errorf("Error saving commit: %w", err)
	}
	e.Commits = append(e.Commits, com)
	if err := e.Save(storage); err != nil {
		return com, err
	}
	return com, nil
}

func (e *Experiment) LatestCommit() *commit.Commit {
	if len(e.Commits) == 0 {
		return nil
	}
	return e.Commits[len(e.Commits)-1]
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

func (e *Experiment) IsRunning() bool {
	now := time.Now()
	lastTolerableHeartbeat := now.Add(-heartbeatRefreshInterval * time.Duration(heartbeatMissTolerance))
	return e.LastHeartbeat.After(lastTolerableHeartbeat)
}

func List(store storage.Storage) ([]*Experiment, error) {
	paths := []string{}
	results := make(chan storage.ListResult)
	go store.MatchFilenamesRecursive(results, "experiments", "replicate-metadata.json")
	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		paths = append(paths, result.Path)
	}
	results = make(chan storage.ListResult)
	go store.MatchFilenamesRecursive(results, "experiments", "replicate-heartbeat.json")
	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		paths = append(paths, result.Path)
	}

	datas, err := store.GetMultiple(paths)
	if err != nil {
		return nil, err
	}

	// First, build experiments from replicate-metadata.json
	experiments := map[string]*Experiment{}
	for path, data := range datas {
		if filepath.Base(path) == "replicate-metadata.json" {
			exp := new(Experiment)
			if err := json.Unmarshal(data, exp); err != nil {
				return nil, fmt.Errorf("Parse error: %s", err)
			}
			experiments[exp.ID] = exp
		}
	}

	// Next, attach heartbeats from replicate-heartbeat.json
	for path, data := range datas {
		if filepath.Base(path) == "replicate-heartbeat.json" {
			heartbeat := new(Heartbeat)
			if err := json.Unmarshal(data, heartbeat); err != nil {
				return nil, fmt.Errorf("Parse error: %s", err)
			}
			if _, ok := experiments[heartbeat.ExperimentID]; !ok {
				console.Warn("Failed to load heartbeat from %q, could not find corresponding experiment %q", path, heartbeat.ExperimentID)
				continue
			}
			experiments[heartbeat.ExperimentID].LastHeartbeat = heartbeat.LastHeartbeat
		}
	}

	experimentList := make([]*Experiment, 0, len(experiments))
	for _, exp := range experiments {
		experimentList = append(experimentList, exp)
	}

	return experimentList, nil
}
