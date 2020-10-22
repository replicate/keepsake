package project

import (
	"encoding/json"
	"path"
	"sort"
	"time"

	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

// Experiment represents a training run
type Experiment struct {
	ID          string         `json:"id"`
	Created     time.Time      `json:"created"`
	Params      param.ValueMap `json:"params"`
	Host        string         `json:"host"`
	User        string         `json:"user"`
	Config      *config.Config `json:"config"`
	Command     string         `json:"command"`
	Path        string         `json:"path"`
	Checkpoints []*Checkpoint  `json:"checkpoints"`
}

type NamedParam struct {
	Name  string
	Value *param.Value
}

// NewExperiment creates an experiment, setting ID and Created
// TODO(andreas): can we get rid of this function?
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

func (c *Experiment) SortedParams() []*NamedParam {
	ret := []*NamedParam{}
	for k, v := range c.Params {
		ret = append(ret, &NamedParam{Name: k, Value: v})
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})
	return ret
}

func (e *Experiment) ShortID() string {
	return e.ID[:7]
}

func (e *Experiment) MetadataPath() string {
	return "metadata/experiments/" + e.ID + ".json"
}

func (e *Experiment) HeartbeatPath() string {
	return "metadata/heartbeats/" + e.ID + ".json"
}

func (e *Experiment) StorageTarPath() string {
	return "experiments/" + e.ID + ".tar.gz"
}

// LatestCheckpoint returns the latest checkpoint for an experiment
func (e *Experiment) LatestCheckpoint() *Checkpoint {
	if len(e.Checkpoints) == 0 {
		return nil
	}
	checkpoints := copyCheckpoints(e.Checkpoints)
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Created.Before(checkpoints[j].Created)
	})
	return checkpoints[len(checkpoints)-1]
}

// BestCheckpoint returns the best checkpoint for an experiment
// according to the primary metric, or nil if primary metric is not
// defined or if none of the checkpoints have the primary metric defined
func (e *Experiment) BestCheckpoint() *Checkpoint {
	if len(e.Checkpoints) == 0 {
		return nil
	}
	checkpoints := copyCheckpoints(e.Checkpoints)

	// Use primary metric from first checkpoint
	// TODO (bfirsh): warn if primary metric differs across checkpoints
	primaryMetric := checkpoints[0].PrimaryMetric
	if primaryMetric == nil {
		return nil
	}

	sort.Slice(checkpoints, func(i, j int) bool {
		iVal, iOK := checkpoints[i].Metrics[primaryMetric.Name]
		jVal, jOK := checkpoints[j].Metrics[primaryMetric.Name]
		if !iOK {
			return true
		}
		if !jOK {
			return false
		}
		if primaryMetric.Goal == GoalMaximize {
			less, err := iVal.LessThan(jVal)
			if err != nil {
				console.Warn("Got error when comparing metrics: %s", err)
			}
			return less
		} else {
			greater, err := iVal.GreaterThan(jVal)
			if err != nil {
				console.Warn("Got error when comparing metrics: %s", err)
			}
			return greater
		}
	})
	best := checkpoints[len(checkpoints)-1]

	// if the last (best) checkpoint in the sorted list doesn't have
	// a value for the primary metric, none of them do
	if _, ok := best.Metrics[primaryMetric.Name]; !ok {
		return nil
	}
	return best
}

func listExperiments(store storage.Storage) ([]*Experiment, error) {
	paths, err := store.List("metadata/experiments/")
	if err != nil {
		return nil, err
	}
	experiments := []*Experiment{}
	for _, p := range paths {
		exp := new(Experiment)
		if err := loadFromPath(store, p, exp); err == nil {
			experiments = append(experiments, exp)
		} else {
			// TODO: should this just be ignored? can this be recovered from?
			console.Warn("Failed to load metadata from %q: %s", p, err)
		}
	}
	return experiments, nil
}

func copyCheckpoints(checkpoints []*Checkpoint) []*Checkpoint {
	copied := make([]*Checkpoint, len(checkpoints))
	copy(copied, checkpoints)
	return copied
}
