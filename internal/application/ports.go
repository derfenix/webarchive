// Package application holds the use cases of the archive and the ports they
// need. Adapters on the outside implement those ports; nothing here imports an
// adapter, a config struct or a transport.
package application

import (
	"context"

	"github.com/google/uuid"

	"github.com/derfenix/webarchive/internal/domain"
)

// PageRepository stores pages and the files of their results.
type PageRepository interface {
	Save(ctx context.Context, page *domain.Page) error
	Get(ctx context.Context, id uuid.UUID) (*domain.Page, error)
	List(ctx context.Context) ([]*domain.Page, error)
	GetFile(ctx context.Context, pageID, fileID uuid.UUID) (*domain.File, error)
	ListPending(ctx context.Context) ([]*domain.Page, error)
}

// Fetched is the result of the single upstream request a page is fetched with:
// its meta and the content the formats reuse.
type Fetched struct {
	Meta   domain.Meta
	Source domain.Source
}

// Archiver fetches a page and archives it in the requested format.
type Archiver interface {
	Fetch(ctx context.Context, pageURL string) (Fetched, error)
	Archive(ctx context.Context, format domain.Format, source domain.Source) ([]domain.File, error)
}

// TaskPublisher hands a page over to the worker.
type TaskPublisher interface {
	Push(ctx context.Context, task ArchiveTask) error
}

// TaskSource is the worker's side of the queue. Pop blocks until a task is
// available or the context is cancelled; it reports false when the queue is
// closed or the context is done.
type TaskSource interface {
	Pop(ctx context.Context) (ArchiveTask, bool)
}

// ArchiveTask is a page to process. Source carries the content already fetched
// when the page was added, so formats do not download it again.
type ArchiveTask struct {
	Page   *domain.Page
	Source domain.Source
}
