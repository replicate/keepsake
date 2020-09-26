package project

import (
	"encoding/json"
	"sort"
	"time"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

type MetricGoal string

const (
	GoalMaximize MetricGoal = "maximize"
	GoalMinimize MetricGoal = "minimize"
)

type PrimaryMetric struct {
	Name string     `json:"name"`
	Goal MetricGoal `json:"goal"`
}

// Checkpoint is a snapshot of an experiment's filesystem
type Checkpoint struct {
	ID            string                  `json:"id"`
	Created       time.Time               `json:"created"`
	ExperimentID  string                  `json:"experiment_id"`
	Metrics       map[string]*param.Value `json:"metrics"`
	Step          int                     `json:"step"`
	Path          string                  `json:"path"`
	PrimaryMetric *PrimaryMetric          `json:"primary_metric"`
}

// NewCheckpoint creates a checkpoint
func NewCheckpoint(experimentID string, metrics map[string]*param.Value) *Checkpoint {
	// FIXME (bfirsh): content addressable (also in Python)
	return &Checkpoint{
		ID:           hash.Random(),
		Created:      time.Now().UTC(),
		ExperimentID: experimentID,
		Metrics:      metrics,
	}
}

// Save a checkpoint, with a copy of the filesystem
func (c *Checkpoint) Save(st storage.Storage, workingDir string) error {
	err := st.PutPath(workingDir, c.StorageDir())
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	return st.Put(c.MetadataPath(), data)
}

func (c *Checkpoint) SortedMetrics() []*NamedParam {
	ret := []*NamedParam{}
	for k, v := range c.Metrics {
		ret = append(ret, &NamedParam{Name: k, Value: v})
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})
	return ret
}

func (c *Checkpoint) ShortID() string {
	return c.ID[:7]
}

func (c *Checkpoint) ShortExperimentID() string {
	return c.ExperimentID[:7]
}

func (c *Checkpoint) MetadataPath() string {
	return "metadata/checkpoints/" + c.ID + ".json"
}

func (c *Checkpoint) StorageDir() string {
	return "checkpoints/" + c.ID
}

func listCheckpoints(store storage.Storage) ([]*Checkpoint, error) {
	paths, err := store.List("metadata/checkpoints/")
	if err != nil {
		return nil, err
	}
	checkpoints := []*Checkpoint{}
	for _, p := range paths {
		com := new(Checkpoint)
		if err := cachedLoadFromPath(store, p, com); err == nil {
			checkpoints = append(checkpoints, com)
		} else {
			console.Warn("Failed to load metadata from %q: %s", p, err)
		}
	}
	return checkpoints, nil
}
