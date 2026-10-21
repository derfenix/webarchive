// Package queue carries archive tasks from the transport to the worker.
package queue

import (
	"context"
	"fmt"

	"github.com/derfenix/webarchive/internal/application"
)

// Memory is an in-process queue. Its buffer lets the transport hand pages over
// without waiting for the worker to pick them up.
func NewMemory(size int) *Memory {
	return &Memory{tasks: make(chan application.ArchiveTask, size)}
}

type Memory struct {
	tasks chan application.ArchiveTask
}

func (m *Memory) Push(ctx context.Context, task application.ArchiveTask) error {
	select {
	case m.tasks <- task:
		return nil

	case <-ctx.Done():
		return fmt.Errorf("push task: %w", ctx.Err())
	}
}

func (m *Memory) Pop(ctx context.Context) (application.ArchiveTask, bool) {
	select {
	case task, open := <-m.tasks:
		return task, open

	case <-ctx.Done():
		return application.ArchiveTask{}, false
	}
}

func (m *Memory) Close() error {
	close(m.tasks)

	return nil
}
