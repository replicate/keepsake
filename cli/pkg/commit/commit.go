package commit

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
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

func ListCommits(store storage.Storage) ([]*Commit, error) {
	paths, err := store.List("metadata/commits/")
	if err != nil {
		return nil, err
	}
	commits := []*Commit{}
	for _, p := range paths {
		com, err := loadCommitFromPath(store, p)
		if err == nil {
			commits = append(commits, com)
		} else {
			console.Warn("Failed to load metadata from %q: %s", p, err)
		}
	}
	return commits, nil
}

func loadCommitFromPath(store storage.Storage, path string) (*Commit, error) {
	com := new(Commit)
	contents, err := store.Get(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(contents, com); err != nil {
		return nil, fmt.Errorf("Parse error: %s", err)
	}
	return com, nil
}

// CommitIDFromPrefix returns the full commit ID given a prefix
func CommitIDFromPrefix(store storage.Storage, prefix string) (string, error) {
	// TODO(andreas): this is a naive implementation, pending data refactoring
	// TODO(bfirsh): fail if the prefix is ambiguous
	commits, err := ListCommits(store)
	if err != nil {
		return "", err
	}
	for _, com := range commits {
		if strings.HasPrefix(com.ID, prefix) {
			return com.ID, nil
		}
	}
	return "", fmt.Errorf("Commit not found: %s", prefix)
}
