package entity

import (
	"context"
	"sync"

	"go.uber.org/zap"
)

type Pages interface {
	Save(ctx context.Context, page *Page) error
}

func NewWorker(ch chan *Page, pages Pages, processor Processor, log *zap.Logger) *Worker {
	return &Worker{pages: pages, processor: processor, log: log, ch: ch}
}

type Worker struct {
	ch        chan *Page
	pages     Pages
	processor Processor
	log       *zap.Logger
}

func (w *Worker) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	w.log.Info("starting")

	for {
		select {
		case <-ctx.Done():
			return

		case page, open := <-w.ch:
			if !open {
				w.log.Warn("channel closed")
				return
			}

			log := w.log.With(zap.Stringer("page_id", page.ID), zap.String("page_url", page.URL))

			log.Info("got new page")

			wg.Add(1)
			go w.do(ctx, wg, page, log)
		}
	}
}

func (w *Worker) do(ctx context.Context, wg *sync.WaitGroup, page *Page, log *zap.Logger) {
	defer wg.Done()

	page.Process(ctx, w.processor)

	log.Debug("page processed")

	if err := w.pages.Save(ctx, page); err != nil {
		w.log.Error(
			"failed to save processed page",
			zap.String("page_id", page.ID.String()),
			zap.String("page_url", page.URL),
			zap.Error(err),
		)
	}
}
