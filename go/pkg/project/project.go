package project

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net"
	"os/user"
	"strings"
	"time"

	"github.com/replicate/replicate/go/pkg/config"
	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/errors"
	"github.com/replicate/replicate/go/pkg/global"
	"github.com/replicate/replicate/go/pkg/param"
	"github.com/replicate/replicate/go/pkg/repository"
)

const IDLength = 64

// Project is essentially a data access object for retrieving
// metadata objects
type Project struct {
	repository        repository.Repository
	directory         string
	experimentsByID   map[string]*Experiment
	heartbeatsByExpID map[string]*Heartbeat
	hasLoaded         bool
}

func NewProject(repo repository.Repository, directory string) *Project {
	return &Project{
		repository: repo,
		directory:  directory,
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

// TODO(andreas): docstring
func (p *Project) ExperimentFromPrefix(prefix string) (*Experiment, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}

	matches := []*Experiment{}

	for id := range p.experimentsByID {
		exp := p.experimentsByID[id]
		if strings.HasPrefix(id, prefix) {
			matches = append(matches, exp)
		}
	}

	if len(matches) == 0 {
		return nil, errors.DoesNotExist("Experiment not found: " + prefix)
	}
	if len(matches) > 1 {
		return nil, fmt.Errorf("Prefix is ambiguous: %s (%d matching experiments)", prefix, len(matches))
	}
	return matches[0], nil
}

// TODO(andreas): docstring
func (p *Project) ExperimentByID(id string) (*Experiment, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, err
	}
	if exp, ok := p.experimentsByID[id]; ok {
		return exp, nil
	}
	return nil, fmt.Errorf("Experiment not found: %s", id)
}

// TODO(andreas): docstring
func (p *Project) CheckpointFromPrefix(prefix string) (*Checkpoint, *Experiment, error) {
	if err := p.ensureLoaded(); err != nil {
		return nil, nil, err
	}

	type match struct {
		checkpoint *Checkpoint
		experiment *Experiment
	}

	matches := []match{}

	for id := range p.experimentsByID {
		exp := p.experimentsByID[id]
		for _, checkpoint := range exp.Checkpoints {
			if strings.HasPrefix(checkpoint.ID, prefix) {
				matches = append(matches, match{
					checkpoint: checkpoint,
					experiment: exp,
				})
			}
		}
	}

	if len(matches) == 0 {
		return nil, nil, fmt.Errorf("Checkpoint not found: %s", prefix)
	}
	if len(matches) > 1 {
		return nil, nil, fmt.Errorf("Prefix is ambiguous: %s (%d matching checkpoints)", prefix, len(matches))
	}

	m := matches[0]
	return m.checkpoint, m.experiment, nil
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

func (p *Project) DeleteCheckpoint(chk *Checkpoint) error {
	if err := p.repository.Delete(chk.StorageTarPath()); err != nil {
		console.Warn("Failed to delete checkpoint storage directory %s: %s", chk.StorageTarPath(), err)
	}
	p.invalidateCache()
	return nil
}

func (p *Project) DeleteExperiment(exp *Experiment) error {
	console.Debug("Deleting experiment: %s", exp.ShortID())
	if err := p.repository.Delete(exp.HeartbeatPath()); err != nil {
		console.Warn("Failed to delete heartbeat file %s: %s", exp.HeartbeatPath(), err)
	}
	if err := p.repository.Delete(exp.StorageTarPath()); err != nil {
		console.Warn("Failed to delete experiment storage directory %s: %s", exp.StorageTarPath(), err)
	}
	if err := p.repository.Delete(exp.MetadataPath()); err != nil {
		console.Warn("Failed to delete experiment metadata file %s: %s", exp.MetadataPath(), err)
	}
	p.invalidateCache()
	return nil
}

type CreateExperimentArgs struct {
	Path           string
	Command        string
	Params         map[string]param.Value
	PythonPackages map[string]string
}

func (p *Project) CreateExperiment(args CreateExperimentArgs, async bool, workChan chan func() error) (*Experiment, error) {
	spec, err := repository.LoadSpec(p.repository)
	if err != nil {
		return nil, err
	}
	if spec == nil {
		if err := repository.WriteSpec(p.repository); err != nil {
			return nil, err
		}
	} else if spec.Version > repository.Version {
		return nil, errors.IncompatibleRepositoryVersion(p.repository.RootURL())
	}

	hostIP, err := localIP()
	host := ""
	if err == nil {
		host = hostIP.String()
	} else {
		console.Warn("Failed to determine host IP: %s", err)
	}
	currentUser, err := user.Current()
	username := ""
	if err == nil {
		username = currentUser.Username
	} else {
		console.Warn("Failed to determine username: %s", err)
	}
	conf := &config.Config{Repository: p.repository.RootURL()}

	exp := &Experiment{
		ID:               generateRandomID(),
		Created:          time.Now().UTC(),
		Params:           args.Params,
		Host:             host,
		User:             username,
		Config:           conf,
		Command:          args.Command,
		Path:             args.Path,
		PythonPackages:   args.PythonPackages,
		ReplicateVersion: global.Version,
	}

	// save json synchronously to uncover repository write issues
	if _, err := p.SaveExperiment(exp); err != nil {
		return nil, err
	}

	work := func() error { return nil }
	if exp.Path != "" {
		work = func() error {
			return p.repository.PutPathTar(p.directory, exp.StorageTarPath(), exp.Path)
		}
	}

	if async {
		workChan <- work
	} else {
		if err := work(); err != nil {
			return nil, err
		}
	}
	return exp, nil
}

type CreateCheckpointArgs struct {
	Path          string
	Step          int64
	Metrics       map[string]param.Value
	PrimaryMetric *PrimaryMetric
}

func (p *Project) CreateCheckpoint(args CreateCheckpointArgs, async bool, workChan chan func() error, quiet bool) (*Checkpoint, error) {
	chk := &Checkpoint{
		ID:            generateRandomID(),
		Created:       time.Now().UTC(),
		Metrics:       args.Metrics,
		Step:          args.Step,
		Path:          args.Path,
		PrimaryMetric: args.PrimaryMetric,
	}

	// if path is empty (i.e. it was None in python), just return
	// the checkpoint without saving anything
	if chk.Path == "" {
		return chk, nil
	}

	work := func() error {
		if err := p.repository.PutPathTar(p.directory, chk.StorageTarPath(), chk.Path); err != nil {
			return err
		}
		if !quiet {
			console.Info("Copied the files from checkpoint %s to %s", chk.ID, chk.StorageTarPath())
		}
		return nil
	}
	if async {
		workChan <- work
	} else {
		if err := work(); err != nil {
			return nil, err
		}
	}

	return chk, nil
}

func (p *Project) SaveExperiment(exp *Experiment) (*Experiment, error) {
	if err := exp.Save(p.repository); err != nil {
		return nil, err
	}
	p.invalidateCache()
	return exp, nil
}

func (p *Project) RefreshHeartbeat(experimentID string) error {
	return CreateHeartbeat(p.repository, experimentID, time.Now().UTC())
}

func (p *Project) StopExperiment(experimentID string) error {
	if err := DeleteHeartbeat(p.repository, experimentID); err != nil {
		return err
	}
	p.invalidateCache()
	return nil
}

func (p *Project) invalidateCache() {
	p.hasLoaded = false
}

// ensureLoaded eagerly loads all the metadata for this project.
// This is highly inefficient, see https://github.com/replicate/replicate/issues/305
func (p *Project) ensureLoaded() error {
	// TODO(andreas): 5(?) second caching instead
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

func generateRandomID() string {
	chars := []rune("0123456789abcdef")
	var b strings.Builder
	for i := 0; i < IDLength; i++ {
		_, err := b.WriteRune(chars[rand.Intn(len(chars))])
		if err != nil {
			// should never happen!
			panic(err)
		}
	}
	return b.String()
}

// Get local IP address of this machine
func localIP() (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, iface := range ifaces {
		addrs, err := iface.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				return v.IP, nil
			case *net.IPAddr:
				return v.IP, nil
			}
		}
	}
	return nil, fmt.Errorf("No local IP address found")
}
