package experiment

import (
	"time"
)

type Heartbeat struct {
	ExperimentID  string    `json:"experiment_id"`
	LastHeartbeat time.Time `json:"last_heartbeat"`
}
