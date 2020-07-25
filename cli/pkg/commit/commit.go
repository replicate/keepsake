package commit

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

// Commit is a snapshot of an experiment's filesystem
type Commit struct {
	ID           string    `json:"id"`
	Created      time.Time `json:"created"`
	ExperimentID string    `json:"experimentID"`

	// TODO(andreas): rename metrics to something else or split it up semantically
	Metrics map[string]*param.Value `json:"metrics"`
}

// NewCommit creates a commit
func NewCommit(experimentID string, metrics map[string]*param.Value) *Commit {
	// FIXME (bfirsh): content addressable (also in Python)
	return &Commit{
		ID:           hash.Random(),
		ExperimentID: experimentID,
		Created:      time.Now().UTC(),
		Metrics:      metrics,
	}
}

// Save a commit, with a copy of the filesystem
func (c *Commit) Save(st storage.Storage, workingDir string) error {
	err := st.PutDirectory(path.Join("commits", c.ID), workingDir)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return err
	}
	return st.Put(path.Join("commits", c.ID, "replicate-metadata.json"), data)
}

func ListCommits(store storage.Storage) ([]*Commit, error) {
	commits := []*Commit{}

	results := make(chan storage.ListResult)
	go store.MatchFilenamesRecursive(results, "commits", "replicate-metadata.json")
	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		com, err := loadCommitFromPath(store, result.Path)
		if err == nil {
			commits = append(commits, com)
		} else {
			console.Warn("Failed to load metadata from %s, got error: %s", result.Path, err)
		}
	}

	return commits, nil
}

func loadCommitFromPath(store storage.Storage, path string) (*Commit, error) {
	contents, err := store.Get(path)
	if err != nil {
		return nil, err
	}
	com := new(Commit)
	if err := json.Unmarshal(contents, com); err != nil {
		return nil, fmt.Errorf("Parse error: %s", err)
	}
	return com, nil
}
