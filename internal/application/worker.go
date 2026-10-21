package application

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

// NewWorker builds the component that turns queued tasks into archived pages.
func NewWorker(service *Service, pages PageRepository, tasks TaskSource, log *zap.Logger) *Worker {
	return &Worker{service: service, pages: pages, tasks: tasks, log: log}
}

type Worker struct {
	service *Service
	pages   PageRepository
	tasks   TaskSource
	log     *zap.Logger
}

// Run processes the pages left unfinished by a previous run and then every task
// from the queue until the context is cancelled. It returns after the in-flight
// pages are done, so their results are stored before the process exits.
func (w *Worker) Run(ctx context.Context) error {
	w.log.Info("starting")

	var wg sync.WaitGroup

	pending, err := w.pages.ListPending(ctx)
	if err != nil {
		w.log.Error("failed to get unprocessed pages", zap.Error(err))
	} else {
		for i := range pending {
			wg.Add(1)

			go func(task ArchiveTask) {
				defer wg.Done()

				w.service.ProcessPage(ctx, task)
			}(ArchiveTask{Page: pending[i]})
		}
	}

	for {
		task, open := w.tasks.Pop(ctx)
		if !open {
			break
		}

		wg.Add(1)

		go func(task ArchiveTask) {
			defer wg.Done()

			w.service.ProcessPage(ctx, task)
		}(task)
	}

	wg.Wait()

	w.log.Info("stopped")

	return nil
}
