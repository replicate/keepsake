package concurrency

import (
	"context"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

type WorkerQueue struct {
	group *errgroup.Group
	ctx   context.Context
	sem   *semaphore.Weighted
}

func NewWorkerQueue(ctx context.Context, maxWorkers int) *WorkerQueue {
	wq := &WorkerQueue{}
	wq.group, wq.ctx = errgroup.WithContext(ctx)
	wq.sem = semaphore.NewWeighted(int64(maxWorkers))
	return wq
}

// Go starts a Go routine as a worker. If there are already maxWorkers running, it
// will block until one of them finishes
//
// Remember to redefine variables in a loop before calling Go(). See pkg/repository/sync.go for an example.
func (wq *WorkerQueue) Go(f func() error) error {
	// Context has error, so let it fall through to Wait(), otherwise
	// sem.Acquire will return error context.Cancelled
	if wq.ctx.Err() != nil {
		return nil
	}
	if err := wq.sem.Acquire(wq.ctx, 1); err != nil {
		return err
	}
	wq.group.Go(func() error {
		defer wq.sem.Release(1)
		return f()
	})
	return nil
}

// Wait until all workers have finished their work. Any errors returned by workers will
// be returned by this function.
func (wq *WorkerQueue) Wait() error {
	return wq.group.Wait()
}
