package commit

import (
	"encoding/json"
	"fmt"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

type Commit struct {
	ID         string                `json:"id"`
	Timestamp  float64               `json:"timestamp"`
	Experiment experiment.Experiment `json:"experiment"`

	// TODO(andreas): rename metrics to something else or split it up semantically
	Metrics map[string]*param.Value `json:"metrics"`
}

func ListCommits(store storage.Storage) ([]*Commit, error) {
	results := make(chan storage.ListResult)
	go store.MatchFilenamesRecursive(results, "commits", "replicate-metadata.json")

	commits := []*Commit{}
	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		com, err := LoadFromPath(store, result.Path)
		if err == nil {
			commits = append(commits, com)
		} else {
			console.Warn("Failed to commit metadata from %s, got error: %s", result.Path, err)
		}
	}
	return commits, nil
}

func LoadFromPath(store storage.Storage, path string) (*Commit, error) {
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
