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
	ID      string                  `json:"id"`
	Created time.Time               `json:"created"`
	Params  map[string]*param.Value `json:"params"`
	Host    string                  `json:"host"`
	User    string                  `json:"user"`
	Config  *config.Config          `json:"config"`
	Command string                  `json:"command"`
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

func (e *Experiment) HasMetrics() bool {
	// TODO(andreas): we should weed out all the circumstances where e.Config or e.Config.Metrics can be nil
	return e.Config != nil && e.Config.Metrics != nil && len(e.Config.Metrics) > 0
}

func (e *Experiment) ShortID() string {
	return e.ID[:7]
}

func listExperiments(store storage.Storage) ([]*Experiment, error) {
	paths, err := store.List("metadata/experiments/")
	if err != nil {
		return nil, err
	}
	experiments := []*Experiment{}
	for _, p := range paths {
		exp := new(Experiment)
		if err := cachedLoadFromPath(store, p, exp); err == nil {
			experiments = append(experiments, exp)
		} else {
			// TODO: should this just be ignored? can this be recovered from?
			console.Warn("Failed to load metadata from %q: %s", p, err)
		}
	}
	return experiments, nil
}
