package project

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/storage"
)

// Project is essentially a data access object for retrieving
// metadata objects
type Project struct {
	store              storage.Storage
	checkpointsByID    map[string]*Checkpoint
	experimentsByID    map[string]*Experiment
	heartbeatsByExpID  map[string]*Heartbeat
	checkpointsByExpID map[string][]*Checkpoint
	hasLoaded          bool
}

func NewProject(store storage.Storage) *Project {
	return &Project{
		store:     store,
		hasLoaded: false,
	}
}

// ExperimentByID returns a particular experiment by ID
func (p *Project) ExperimentByID(experimentID string) (*Experiment, error) {
	// TODO(andreas): right now we naively load all experiments in the project. this could be improved with local indexing
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	exp, ok := p.experimentsByID[experimentID]
	if !ok {
		return nil, fmt.Errorf("No experiment found with ID %s", experimentID)
	}
	return exp, nil
}

// Experiments returns all experiments in this project
func (p *Project) Experiments() ([]*Experiment, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	experiments := []*Experiment{}
	for _, exp := range p.experimentsByID {
		experiments = append(experiments, exp)
	}
	return experiments, nil
}

// ExperimentCheckpoints returns all checkpoints for a particular experiment
func (p *Project) ExperimentCheckpoints(experimentID string) ([]*Checkpoint, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	checkpoints, ok := p.checkpointsByExpID[experimentID]
	if !ok {
		return []*Checkpoint{}, nil
	}
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Created.Before(checkpoints[j].Created)
	})
	return checkpoints, nil
}

// ExperimentLatestCheckpoint returns the latest checkpoint for an experiment
func (p *Project) ExperimentLatestCheckpoint(experimentID string) (*Checkpoint, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	checkpoints, ok := p.checkpointsByExpID[experimentID]
	if !ok || len(checkpoints) == 0 {
		return nil, nil
	}
	checkpoints = copyCheckpoints(checkpoints)
	sort.Slice(checkpoints, func(i, j int) bool {
		return checkpoints[i].Created.Before(checkpoints[j].Created)
	})
	return checkpoints[len(checkpoints)-1], nil
}

// ExperimentBestCheckpoint returns the best checkpoint for an experiment
// according to the primary metric, or nil if primary metric is not
// defined or if none of the checkpoints have the primary metric defined
func (p *Project) ExperimentBestCheckpoint(experimentID string) (*Checkpoint, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}

	checkpoints, ok := p.checkpointsByExpID[experimentID]
	if !ok || len(checkpoints) == 0 {
		return nil, nil
	}
	checkpoints = copyCheckpoints(checkpoints)

	// Use primary metric from first checkpoint
	// TODO (bfirsh): warn if primary metric differs across checkpoints
	primaryMetric := checkpoints[0].PrimaryMetric
	if primaryMetric == nil {
		return nil, nil
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
		return nil, nil
	}

	return best, nil
}

func (p *Project) ExperimentIsRunning(experimentID string) (bool, error) {
	if err := p.ensureLoaded(); err != nil {
		return false, err
	}
	heartbeat, ok := p.heartbeatsByExpID[experimentID]
	if !ok {
		// TODO(bfirsh): unknown state? https://github.com/replicate/replicate/issues/36
		console.Debug("No heartbeat found for experiment %s", experimentID)
		return false, nil
	}
	return heartbeat.IsRunning(), nil
}

// CheckpointFromPrefix returns a checkpoint given the prefix of the checkpoint ID
func (p *Project) CheckpointFromPrefix(prefix string) (*Checkpoint, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	matchingIDs := p.checkpointIDsFromPrefix(prefix)
	if len(matchingIDs) == 0 {
		return nil, fmt.Errorf("Checkpoint not found: %s", prefix)
	}
	if len(matchingIDs) > 1 {
		return nil, fmt.Errorf("Prefix is ambiguous: %s (%d matching checkpoints)", prefix, len(matchingIDs))
	}
	return p.checkpointsByID[matchingIDs[0]], nil
}

type CheckpointOrExperiment struct {
	Checkpoint *Checkpoint
	Experiment *Experiment
}

// CheckpointOrExperimentFromPrefix returns a checkpoint/experiment given a
// prefix. This is a single function so we can detect ambiguities
// across both checkpoints and experiments.
func (p *Project) CheckpointOrExperimentFromPrefix(prefix string) (*CheckpointOrExperiment, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	matchingCheckpointIDs := p.checkpointIDsFromPrefix(prefix)
	matchingExperimentIDs := p.experimentIDsFromPrefix(prefix)
	numMatches := len(matchingCheckpointIDs) + len(matchingExperimentIDs)
	if numMatches == 0 {
		return nil, fmt.Errorf("Checkpoint/experiment not found: %s", prefix)
	}
	if numMatches > 1 {
		return nil, fmt.Errorf("Prefix is ambiguous: %s (%d matching checkpoints/experiments)", prefix, numMatches)
	}
	if len(matchingCheckpointIDs) == 1 {
		return &CheckpointOrExperiment{Checkpoint: p.checkpointsByID[matchingCheckpointIDs[0]]}, nil
	}
	return &CheckpointOrExperiment{Experiment: p.experimentsByID[matchingExperimentIDs[0]]}, nil
}

func (p *Project) checkpointIDsFromPrefix(prefix string) []string {
	if !p.hasLoaded {
		panic("Logical error, project has not loaded")
	}
	matchingIDs := []string{}
	for id := range p.checkpointsByID {
		if strings.HasPrefix(id, prefix) {
			matchingIDs = append(matchingIDs, id)
		}
	}
	return matchingIDs
}

func (p *Project) DeleteCheckpoint(com *Checkpoint) error {
	if err := p.store.Delete(com.StorageDir()); err != nil {
		// TODO(andreas): return err if com.StorageDir() exists but some other error occurs
		console.Warn("Failed to delete checkpoint storage directory %s: %s", com.StorageDir(), err)
	}
	if err := p.store.Delete(com.MetadataPath()); err != nil {
		// TODO(andreas): return err if com.MetadataPath() exists but some other error occurs
		console.Warn("Failed to delete checkpoint metadata file %s: %s", com.MetadataPath(), err)
	}
	return nil
}

func (p *Project) DeleteExperiment(exp *Experiment) error {
	if err := p.store.Delete(exp.HeartbeatPath()); err != nil {
		// TODO(andreas): return err if exp.HeartbeatPath() exists but some other error occurs
		console.Warn("Failed to delete heartbeat file %s: %s", exp.HeartbeatPath(), err)
	}
	if err := p.store.Delete(exp.StorageDir()); err != nil {
		// TODO(andreas): return err if com.StorageDir() exists but some other error occurs
		console.Warn("Failed to delete checkpoint storage directory %s: %s", exp.StorageDir(), err)
	}
	if err := p.store.Delete(exp.MetadataPath()); err != nil {
		// TODO(andreas): return err if exp.MetadataPath() exists but some other error occurs
		console.Warn("Failed to delete experiment metadata file %s: %s", exp.MetadataPath(), err)
	}
	return nil
}

func (p *Project) experimentIDsFromPrefix(prefix string) []string {
	if !p.hasLoaded {
		panic("Logical error, project has not loaded")
	}
	matchingIDs := []string{}
	for id := range p.experimentsByID {
		if strings.HasPrefix(id, prefix) {
			matchingIDs = append(matchingIDs, id)
		}
	}
	return matchingIDs
}

// ensureLoaded eagerly loads all the metadata for this project.
// TODO(andreas): this is a naive approach, we should instead use
// some sort of indexing for efficiency.
func (p *Project) ensureLoaded() error {
	if p.hasLoaded {
		return nil
	}
	experiments, err := listExperiments(p.store)
	if err != nil {
		return err
	}
	checkpoints, err := listCheckpoints(p.store)
	if err != nil {
		return err
	}
	heartbeats, err := listHeartbeats(p.store)
	if err != nil {
		heartbeats = []*Heartbeat{}
		console.Warn("Failed to load heartbeats: %s", err)
	}
	p.setObjects(experiments, checkpoints, heartbeats)
	p.hasLoaded = true
	return nil
}

func (p *Project) setObjects(experiments []*Experiment, checkpoints []*Checkpoint, heartbeats []*Heartbeat) {
	p.experimentsByID = map[string]*Experiment{}
	for _, exp := range experiments {
		p.experimentsByID[exp.ID] = exp
	}
	p.checkpointsByID = map[string]*Checkpoint{}
	for _, com := range checkpoints {
		p.checkpointsByID[com.ID] = com
	}
	p.heartbeatsByExpID = map[string]*Heartbeat{}
	for _, hb := range heartbeats {
		p.heartbeatsByExpID[hb.ExperimentID] = hb
	}
	p.checkpointsByExpID = map[string][]*Checkpoint{}
	for _, com := range checkpoints {
		if p.checkpointsByExpID[com.ExperimentID] == nil {
			p.checkpointsByExpID[com.ExperimentID] = []*Checkpoint{}
		}
		p.checkpointsByExpID[com.ExperimentID] = append(p.checkpointsByExpID[com.ExperimentID], com)
	}
}

func copyCheckpoints(checkpoints []*Checkpoint) []*Checkpoint {
	copied := make([]*Checkpoint, len(checkpoints))
	copy(copied, checkpoints)
	return copied
}

func loadFromPath(store storage.Storage, path string, obj interface{}) error {
	contents, err := store.Get(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(contents, obj); err != nil {
		return fmt.Errorf("Parse error: %s", err)
	}
	return nil
}
