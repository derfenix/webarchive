package application

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/derfenix/webarchive/internal/domain"
)

type AddPageInput struct {
	URL         string
	Description string
	Formats     []domain.Format
}

// NewService wires the archive use cases to their ports.
func NewService(pages PageRepository, archiver Archiver, tasks TaskPublisher, log *zap.Logger) *Service {
	return &Service{
		pages:    pages,
		archiver: archiver,
		tasks:    tasks,
		log:      log,
		newID:    uuid.New,
		now:      time.Now,
	}
}

type Service struct {
	pages    PageRepository
	archiver Archiver
	tasks    TaskPublisher
	log      *zap.Logger

	// newID and now are injectable to keep use cases deterministic in tests.
	newID func() uuid.UUID
	now   func() time.Time
}

// AddPage registers a page, fetches its meta and hands it over to the worker.
// The fetched content travels with the task, so formats reuse the response body
// instead of downloading the page again.
func (s *Service) AddPage(ctx context.Context, in AddPageInput) (*domain.Page, error) {
	if in.URL == "" {
		//nolint:staticcheck // the message is returned to API clients as it is
		return nil, &ValidationError{Field: "url", Err: errors.New("Value is required")}
	}

	formats := in.Formats
	if len(formats) == 0 {
		formats = domain.AllFormats()
	}

	page := domain.NewPage(s.newID(), in.URL, in.Description, s.now(), formats...)

	fetched, err := s.archiver.Fetch(ctx, in.URL)
	if err != nil {
		page.SetMetaError(err)
	} else {
		page.SetMeta(fetched.Meta)
	}

	if err := s.pages.Save(ctx, page); err != nil {
		return nil, fmt.Errorf("save page: %w", err)
	}

	// A page that was not handed over stays "new" and is picked up again on the
	// next start, so a failed push does not lose the request.
	if err := s.tasks.Push(ctx, ArchiveTask{Page: page, Source: fetched.Source}); err != nil {
		return nil, fmt.Errorf("enqueue page: %w", err)
	}

	return page, nil
}

func (s *Service) GetPage(ctx context.Context, id uuid.UUID) (*domain.Page, error) {
	page, err := s.pages.Get(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("get page: %w", err)
	}

	return page, nil
}

func (s *Service) ListPages(ctx context.Context) ([]*domain.Page, error) {
	pages, err := s.pages.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list pages: %w", err)
	}

	return pages, nil
}

func (s *Service) GetFile(ctx context.Context, pageID, fileID uuid.UUID) (*domain.File, error) {
	file, err := s.pages.GetFile(ctx, pageID, fileID)
	if err != nil {
		return nil, fmt.Errorf("get file: %w", err)
	}

	return file, nil
}

// ProcessPage archives every format of the page and stores the outcome. The
// page is saved twice on purpose: as "processing" first, so a restart knows the
// page was in flight, and with the results after.
func (s *Service) ProcessPage(ctx context.Context, task ArchiveTask) {
	page := task.Page

	log := s.log.With(zap.Stringer("page_id", page.ID), zap.String("page_url", page.URL))

	page.StartProcessing()

	if err := s.pages.Save(ctx, page); err != nil {
		log.Error("failed to save processing page", zap.Error(err))
	}

	results := s.archiveFormats(ctx, page, task.Source)

	page.Finish(results)

	log.Debug("page processed")

	if err := s.pages.Save(ctx, page); err != nil {
		log.Error("failed to save processed page", zap.Error(err))
	}
}

func (s *Service) archiveFormats(ctx context.Context, page *domain.Page, source domain.Source) []domain.Result {
	results := make([]domain.Result, len(page.Formats))

	var wg sync.WaitGroup

	wg.Add(len(page.Formats))

	// Every goroutine writes its own index: results needs no lock.
	for i, format := range page.Formats {
		go func(i int, format domain.Format) {
			defer wg.Done()

			files, err := s.archiveOne(ctx, format, source)

			results[i] = domain.Result{Format: format, Files: files, Err: err}
		}(i, format)
	}

	wg.Wait()

	return results
}

// archiveOne isolates one format: a panic in a processor fails that format only.
func (s *Service) archiveOne(ctx context.Context, format domain.Format, source domain.Source) (files []domain.File, err error) {
	defer func() {
		if r := recover(); r != nil {
			files = nil
			err = fmt.Errorf("recovered from panic: %v (%s)", r, string(debug.Stack()))
		}
	}()

	files, err = s.archiver.Archive(ctx, format, source)
	if err != nil {
		return nil, fmt.Errorf("process: %w", err)
	}

	return files, nil
}
