package project

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/repository"
)

// Project is essentially a data access object for retrieving
// metadata objects
type Project struct {
	repository        repository.Repository
	experimentsByID   map[string]*Experiment
	heartbeatsByExpID map[string]*Heartbeat
	hasLoaded         bool
}

func NewProject(repo repository.Repository) *Project {
	return &Project{
		repository: repo,
		hasLoaded:  false,
	}
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

	matches := []*CheckpointOrExperiment{}
	for id := range p.experimentsByID {
		exp := p.experimentsByID[id]
		if strings.HasPrefix(id, prefix) {
			matches = append(matches, &CheckpointOrExperiment{Experiment: exp})
		}

		for _, checkpoint := range exp.Checkpoints {
			if strings.HasPrefix(checkpoint.ID, prefix) {
				matches = append(matches, &CheckpointOrExperiment{Experiment: exp, Checkpoint: checkpoint})
			}
		}
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("Checkpoint/experiment not found: %s", prefix)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("Prefix is ambiguous: %s (%d matching checkpoints/experiments)", prefix, len(matches))
	}
	return matches[0], nil
}

func (p *Project) DeleteCheckpoint(com *Checkpoint) error {
	if err := p.repository.Delete(com.StorageTarPath()); err != nil {
		console.Warn("Failed to delete checkpoint storage directory %s: %s", com.StorageTarPath(), err)
	}
	return nil
}

func (p *Project) DeleteExperiment(exp *Experiment) error {
	if err := p.repository.Delete(exp.HeartbeatPath()); err != nil {
		console.Warn("Failed to delete heartbeat file %s: %s", exp.HeartbeatPath(), err)
	}
	if err := p.repository.Delete(exp.StorageTarPath()); err != nil {
		console.Warn("Failed to delete checkpoint storage directory %s: %s", exp.StorageTarPath(), err)
	}
	if err := p.repository.Delete(exp.MetadataPath()); err != nil {
		console.Warn("Failed to delete experiment metadata file %s: %s", exp.MetadataPath(), err)
	}
	return nil
}

// ensureLoaded eagerly loads all the metadata for this project.
// TODO(andreas): this is a naive approach, we should instead use
// some sort of indexing for efficiency.
// TODO(bfirsh): loading all metadata into memory on each run is not... great
func (p *Project) ensureLoaded() error {
	if p.hasLoaded {
		return nil
	}
	experiments, err := listExperiments(p.repository)
	if err != nil {
		return err
	}
	heartbeats, err := listHeartbeats(p.repository)
	if err != nil {
		heartbeats = []*Heartbeat{}
		console.Warn("Failed to load heartbeats: %s", err)
	}
	p.setObjects(experiments, heartbeats)
	p.hasLoaded = true
	return nil
}

func (p *Project) setObjects(experiments []*Experiment, heartbeats []*Heartbeat) {
	p.experimentsByID = map[string]*Experiment{}
	for _, exp := range experiments {
		p.experimentsByID[exp.ID] = exp
	}
	p.heartbeatsByExpID = map[string]*Heartbeat{}
	for _, hb := range heartbeats {
		p.heartbeatsByExpID[hb.ExperimentID] = hb
	}
}

func loadFromPath(repo repository.Repository, path string, obj interface{}) error {
	contents, err := repo.Get(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(contents, obj); err != nil {
		return fmt.Errorf("Parse error: %s", err)
	}
	return nil
}
