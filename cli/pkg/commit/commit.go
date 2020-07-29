package commit

import (
	"encoding/json"
	"fmt"
	"path"
	"strings"
	"time"

	"replicate.ai/cli/pkg/cache"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/experiment"
	"replicate.ai/cli/pkg/hash"
	"replicate.ai/cli/pkg/param"
	"replicate.ai/cli/pkg/storage"
)

// Commit is a snapshot of an experiment's filesystem
type Commit struct {
	ID         string                `json:"id"`
	Created    time.Time             `json:"created"`
	Experiment experiment.Experiment `json:"experiment"`

	// TODO(andreas): rename metrics to something else or split it up semantically
	Metrics map[string]*param.Value `json:"metrics"`
}

// NewCommit creates a commit
func NewCommit(experiment experiment.Experiment, metrics map[string]*param.Value) *Commit {
	// FIXME (bfirsh): content addressable (also in Python)
	return &Commit{
		ID:         hash.Random(),
		Created:    time.Now().UTC(),
		Experiment: experiment,
		Metrics:    metrics,
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
	return st.Put(path.Join("commits", c.ID, "replicate-metadata.json"), data)
}

func ListCommits(store storage.Storage) ([]*Commit, error) {
	commits := []*Commit{}
	heartbeatsByExpID := map[string]*experiment.Heartbeat{}

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

	// heartbeats are stored in experiments/<id>/replicate-heartbeat.json
	// we fetch all heartbeats and then attach them to the commits'
	// experiment metadata
	results = make(chan storage.ListResult)
	go store.MatchFilenamesRecursive(results, "experiments", "replicate-heartbeat.json")
	for result := range results {
		if result.Error != nil {
			return nil, result.Error
		}
		heartbeat, err := loadHeartbeatFromPath(store, result.Path)
		if err == nil {
			heartbeatsByExpID[heartbeat.ExperimentID] = heartbeat
		} else {
			console.Warn("Failed to load heartbeat from %s, got error: %s", result.Path, err)
		}
	}
	attachHeartbeatsToCommitExperiments(commits, heartbeatsByExpID)

	return commits, nil
}

func attachHeartbeatsToCommitExperiments(commits []*Commit, heartbeatsByExpID map[string]*experiment.Heartbeat) {
	missingHeartbeats := map[string]bool{}
	for _, commit := range commits {
		heartbeat, ok := heartbeatsByExpID[commit.Experiment.ID]
		if ok {
			commit.Experiment.LastHeartbeat = heartbeat.LastHeartbeat
			commit.Experiment.Running = experiment.IsRunning(heartbeat.LastHeartbeat)
		} else {
			if hasLogged := missingHeartbeats[commit.Experiment.ID]; !hasLogged {
				console.Warn("Heartbeat not found for experiment %s", commit.Experiment.ID)
				missingHeartbeats[commit.Experiment.ID] = true
			}
		}
	}
}

func loadCommitFromPath(store storage.Storage, path string) (*Commit, error) {
	// TODO(andreas): just added caching here to speed up
	// development. Gettin commitID this way is a hack
	commitID := strings.Split(path, "/")[1]
	cacheKey := "commit-" + commitID
	com := new(Commit)
	if ok := cache.GetStruct(cacheKey, com); ok {
		return com, nil
	}

	contents, err := store.Get(path)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(contents, com); err != nil {
		return nil, fmt.Errorf("Parse error: %s", err)
	}
	if err := cache.SetStruct(cacheKey, com); err != nil {
		return nil, err
	}
	return com, nil
}

func loadHeartbeatFromPath(store storage.Storage, path string) (*experiment.Heartbeat, error) {
	contents, err := store.Get(path)
	if err != nil {
		return nil, err
	}
	hb := new(experiment.Heartbeat)
	if err := json.Unmarshal(contents, hb); err != nil {
		return nil, fmt.Errorf("Parse error: %s", err)
	}
	return hb, nil
}

// CommitIDFromPrefix returns the full commit ID given a prefix
func CommitIDFromPrefix(store storage.Storage, prefix string) (string, error) {
	// TODO(andreas): this is a naive implementation, pending data refactoring
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
