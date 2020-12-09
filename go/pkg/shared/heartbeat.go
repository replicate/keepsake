package shared

import (
	"time"

	"github.com/replicate/replicate/go/pkg/console"
	"github.com/replicate/replicate/go/pkg/project"
)

type HeartbeatProcess struct {
	project      *project.Project
	experimentID string
	ticker       *time.Ticker
	done         chan struct{}
}

func StartHeartbeat(proj *project.Project, experimentID string) *HeartbeatProcess {
	h := &HeartbeatProcess{
		project:      proj,
		experimentID: experimentID,
		ticker:       time.NewTicker(5 * time.Second),
		done:         make(chan struct{}),
	}
	go func() {
		for {
			select {
			case <-h.done:
				return
			case <-h.ticker.C:
				h.Refresh()
			}
		}
	}()
	return h
}

func (h *HeartbeatProcess) Refresh() {
	if err := h.project.RefreshHeartbeat(h.experimentID); err != nil {
		console.Error("Failed to refresh heartbeat: %v", err)
	}
}

func (h *HeartbeatProcess) Kill() {
	h.ticker.Stop()
	h.done <- struct{}{}
}
