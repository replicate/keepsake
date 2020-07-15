package gcp

import (
	"context"
	"fmt"
	"time"
)

const refreshInterval = 10 * time.Second

func waitForOperation(ctx context.Context, isDone func() (bool, error)) error {
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("timeout waiting for operation to complete")
		case <-ticker.C:
			done, err := isDone()
			if err != nil {
				return fmt.Errorf("Operation failed: %s", err)
			}

			if done {
				return nil
			}
		}
	}
}
