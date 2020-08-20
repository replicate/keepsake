package project

import (
	"encoding/json"
	"path"
	"sort"
	"time"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

// Commit is a snapshot of an experiment's filesystem
type Commit struct {
	ID           string                  `json:"id"`
	Created      time.Time               `json:"created"`
	ExperimentID string                  `json:"experiment_id"`
	Labels       map[string]*param.Value `json:"labels"`
	Step         int                     `json:"step"`
	Path         string                  `json:"path"`
}

// NewCommit creates a commit
func NewCommit(experimentID string, labels map[string]*param.Value) *Commit {
	// FIXME (bfirsh): content addressable (also in Python)
	return &Commit{
		ID:           hash.Random(),
		Created:      time.Now().UTC(),
		ExperimentID: experimentID,
		Labels:       labels,
	}
}

// Save a commit, with a copy of the filesystem
func (c *Commit) Save(st storage.Storage, workingDir string) error {
	err := st.PutDirectory(workingDir, path.Join("commits", c.ID))
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	return st.Put(path.Join("metadata", "commits", c.ID+".json"), data)
}

func (c *Commit) SortedLabels() []*NamedParam {
	ret := []*NamedParam{}
	for k, v := range c.Labels {
		ret = append(ret, &NamedParam{Name: k, Value: v})
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Name < ret[j].Name
	})
	return ret
}

func (c *Commit) ShortID() string {
	return c.ID[:7]
}

func (c *Commit) ShortExperimentID() string {
	return c.ExperimentID[:7]
}

func (c *Commit) MetadataPath() string {
	return "metadata/commits/" + c.ID + ".json"
}

func (c *Commit) StorageDir() string {
	return "commits/" + c.ID
}

func listCommits(store storage.Storage) ([]*Commit, error) {
	paths, err := store.List("metadata/commits/")
	if err != nil {
		return nil, err
	}
	commits := []*Commit{}
	for _, p := range paths {
		com := new(Commit)
		if err := cachedLoadFromPath(store, p, com); err == nil {
			commits = append(commits, com)
		} else {
			console.Warn("Failed to load metadata from %q: %s", p, err)
		}
	}
	return commits, nil
}
