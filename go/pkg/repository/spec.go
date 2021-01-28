package repository

import (
	"encoding/json"
	"fmt"

	"github.com/replicate/keepsake/go/pkg/errors"
)

const Version = 1
const SpecPath = "repository.json"

type Spec struct {
	Version int `json:"version"`
}

// LoadSpec returns the repository spec, or nil if the repository doesn't have a spec file
func LoadSpec(r Repository) (*Spec, error) {
	raw, err := r.Get(SpecPath)
	if err != nil {
		if errors.IsDoesNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("Failed to read %s/%s: %v", r.RootURL(), SpecPath, err)
	}

	spec := &Spec{}
	if err := json.Unmarshal(raw, spec); err != nil {
		return nil, errors.CorruptedRepositorySpec(r.RootURL(), SpecPath, err)
	}

	return spec, nil
}

func WriteSpec(r Repository) error {
	spec := Spec{Version: Version}
	raw, err := json.Marshal(&spec)
	if err != nil {
		panic(err) // should never happen
	}
	return r.Put(SpecPath, raw)
}
