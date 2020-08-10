package project

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"replicate.ai/cli/pkg/cache"
	"replicate.ai/cli/pkg/config"
	"replicate.ai/cli/pkg/console"
	"replicate.ai/cli/pkg/storage"
)

// Project is essentially a data access object for retrieving
// metadata objects
type Project struct {
	store             storage.Storage
	commitsByID       map[string]*Commit
	experimentsByID   map[string]*Experiment
	heartbeatsByExpID map[string]*Heartbeat
	commitsByExpID    map[string][]*Commit
	hasLoaded         bool
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

// ExperimentCommits returns all commits for a particular experiment
func (p *Project) ExperimentCommits(experimentID string) ([]*Commit, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	commits, ok := p.commitsByExpID[experimentID]
	if !ok {
		return []*Commit{}, nil
	}
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Created.Before(commits[j].Created)
	})
	return commits, nil
}

// ExperimentLatestCommit returns the latest commit for an experiment
func (p *Project) ExperimentLatestCommit(experimentID string) (*Commit, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	commits, ok := p.commitsByExpID[experimentID]
	if !ok || len(commits) == 0 {
		return nil, nil
	}
	commits = copyCommits(commits)
	sort.Slice(commits, func(i, j int) bool {
		return commits[i].Created.Before(commits[j].Created)
	})
	return commits[len(commits)-1], nil
}

// ExperimentBestCommit returns the best commit for an experiment
// according to the primary metric, or nil if primary metric is not
// defined or if none of the commits have the primary metric defined
func (p *Project) ExperimentBestCommit(experimentID string) (*Commit, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	exp, ok := p.experimentsByID[experimentID]
	if !ok {
		return nil, fmt.Errorf("No experiment found with ID %s", experimentID)
	}
	conf := exp.Config
	if conf == nil {
		conf = new(config.Config)
	}

	primaryMetric := conf.PrimaryMetric()
	if primaryMetric == nil {
		return nil, nil
	}
	commits, ok := p.commitsByExpID[experimentID]
	if !ok || len(commits) == 0 {
		return nil, nil
	}
	commits = copyCommits(commits)

	sort.Slice(commits, func(i, j int) bool {
		iVal, iOK := commits[i].Labels[primaryMetric.Name]
		jVal, jOK := commits[j].Labels[primaryMetric.Name]
		if !iOK {
			return true
		}
		if !jOK {
			return false
		}
		if primaryMetric.Goal == config.GoalMaximize {
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
	best := commits[len(commits)-1]

	// if the last (best) commit in the sorted list doesn't have
	// a value for the primary metric, none of them do
	if _, ok := best.Labels[primaryMetric.Name]; !ok {
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

// CommitFromPrefix returns a commit given the prefix of the commit ID
func (p *Project) CommitFromPrefix(prefix string) (*Commit, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	matchingIDs := p.commitIDsFromPrefix(prefix)
	if len(matchingIDs) == 0 {
		return nil, fmt.Errorf("Commit not found: %s", prefix)
	}
	if len(matchingIDs) > 1 {
		return nil, fmt.Errorf("Prefix is ambiguous: %s (%d matching commits)", prefix, len(matchingIDs))
	}
	return p.commitsByID[matchingIDs[0]], nil
}

type CommitOrExperiment struct {
	Commit     *Commit
	Experiment *Experiment
}

// CommitOrExperimentFromPrefix returns a commit/experiment given a
// prefix. This is a single function so we can detect ambiguities
// across both commits and experiments.
func (p *Project) CommitOrExperimentFromPrefix(prefix string) (*CommitOrExperiment, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	matchingCommitIDs := p.commitIDsFromPrefix(prefix)
	matchingExperimentIDs := p.experimentIDsFromPrefix(prefix)
	numMatches := len(matchingCommitIDs) + len(matchingExperimentIDs)
	if numMatches == 0 {
		return nil, fmt.Errorf("Commit/experiment not found: %s", prefix)
	}
	if numMatches > 1 {
		return nil, fmt.Errorf("Prefix is ambiguous: %s (%d matching commits/experiments)", prefix, numMatches)
	}
	if len(matchingCommitIDs) == 1 {
		return &CommitOrExperiment{Commit: p.commitsByID[matchingCommitIDs[0]]}, nil
	}
	return &CommitOrExperiment{Experiment: p.experimentsByID[matchingExperimentIDs[0]]}, nil
}

func (p *Project) commitIDsFromPrefix(prefix string) []string {
	if !p.hasLoaded {
		panic("Logical error, project has not loaded")
	}
	matchingIDs := []string{}
	for id := range p.commitsByID {
		if strings.HasPrefix(id, prefix) {
			matchingIDs = append(matchingIDs, id)
		}
	}
	return matchingIDs
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
	commits, err := listCommits(p.store)
	if err != nil {
		return err
	}
	heartbeats, err := listHeartbeats(p.store)
	if err != nil {
		heartbeats = []*Heartbeat{}
		console.Warn("Failed to load heartbeats: %s", err)
	}
	p.setObjects(experiments, commits, heartbeats)
	p.hasLoaded = true
	return nil
}

func (p *Project) setObjects(experiments []*Experiment, commits []*Commit, heartbeats []*Heartbeat) {
	p.experimentsByID = map[string]*Experiment{}
	for _, exp := range experiments {
		p.experimentsByID[exp.ID] = exp
	}
	p.commitsByID = map[string]*Commit{}
	for _, com := range commits {
		p.commitsByID[com.ID] = com
	}
	p.heartbeatsByExpID = map[string]*Heartbeat{}
	for _, hb := range heartbeats {
		p.heartbeatsByExpID[hb.ExperimentID] = hb
	}
	p.commitsByExpID = map[string][]*Commit{}
	for _, com := range commits {
		if p.commitsByExpID[com.ExperimentID] == nil {
			p.commitsByExpID[com.ExperimentID] = []*Commit{}
		}
		p.commitsByExpID[com.ExperimentID] = append(p.commitsByExpID[com.ExperimentID], com)
	}
}

func copyCommits(commits []*Commit) []*Commit {
	copied := make([]*Commit, len(commits))
	copy(copied, commits)
	return copied
}

func cachedLoadFromPath(store storage.Storage, path string, obj interface{}) error {
	if ok := cache.GetStruct(path, obj); ok {
		return nil
	}
	contents, err := store.Get(path)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(contents, obj); err != nil {
		return fmt.Errorf("Parse error: %s", err)
	}
	return cache.SetStruct(path, obj)
}
