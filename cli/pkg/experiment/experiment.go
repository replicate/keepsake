package experiment

import (
	"replicate.ai/cli/pkg/param"
)

type Experiment struct {
	ID        string                  `json:"id"`
	Timestamp float64                 `json:"timestamp"`
	Params    map[string]*param.Value `json:"params"`
}
