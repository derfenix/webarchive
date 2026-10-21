package application

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/derfenix/webarchive/internal/domain"
)

type fakeRepo struct {
	mu      sync.Mutex
	pages   map[uuid.UUID]*domain.Page
	pending []*domain.Page
	saves   []*domain.Page
	saveErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{pages: make(map[uuid.UUID]*domain.Page)}
}

func (r *fakeRepo) Save(_ context.Context, page *domain.Page) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.saveErr != nil {
		return r.saveErr
	}

	r.pages[page.ID] = page

	// Snapshot: the service keeps mutating the page it holds, and a test needs
	// to see what was stored at each save.
	snapshot := *page
	r.saves = append(r.saves, &snapshot)

	return nil
}

func (r *fakeRepo) Get(_ context.Context, id uuid.UUID) (*domain.Page, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	page, ok := r.pages[id]
	if !ok {
		return nil, domain.ErrNotFound
	}

	return page, nil
}

func (r *fakeRepo) List(context.Context) ([]*domain.Page, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	pages := make([]*domain.Page, 0, len(r.pages))

	for _, page := range r.pages {
		pages = append(pages, page)
	}

	return pages, nil
}

func (r *fakeRepo) GetFile(_ context.Context, pageID, fileID uuid.UUID) (*domain.File, error) {
	page, err := r.Get(context.Background(), pageID)
	if err != nil {
		return nil, err
	}

	for i := range page.Results {
		for j := range page.Results[i].Files {
			if page.Results[i].Files[j].ID == fileID {
				return &page.Results[i].Files[j], nil
			}
		}
	}

	return nil, domain.ErrNotFound
}

func (r *fakeRepo) ListPending(context.Context) ([]*domain.Page, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.pending, nil
}

func (r *fakeRepo) saved() []*domain.Page {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.saves
}

type fakeArchiver struct {
	fetched  Fetched
	fetchErr error

	files      []domain.File
	archiveErr error
	panicWith  any

	mu             sync.Mutex
	archived       []domain.Format
	archivedSource []domain.Source
}

func (a *fakeArchiver) Fetch(context.Context, string) (Fetched, error) {
	if a.fetchErr != nil {
		return Fetched{}, a.fetchErr
	}

	return a.fetched, nil
}

func (a *fakeArchiver) Archive(_ context.Context, format domain.Format, source domain.Source) ([]domain.File, error) {
	a.mu.Lock()
	a.archived = append(a.archived, format)
	a.archivedSource = append(a.archivedSource, source)
	a.mu.Unlock()

	if a.panicWith != nil {
		panic(a.panicWith)
	}

	if a.archiveErr != nil {
		return nil, a.archiveErr
	}

	return a.files, nil
}

func (a *fakeArchiver) formats() []domain.Format {
	a.mu.Lock()
	defer a.mu.Unlock()

	return a.archived
}

type fakePublisher struct {
	tasks []ArchiveTask
	err   error
}

func (p *fakePublisher) Push(_ context.Context, task ArchiveTask) error {
	if p.err != nil {
		return p.err
	}

	p.tasks = append(p.tasks, task)

	return nil
}

func TestService_AddPage(t *testing.T) {
	t.Parallel()

	fixedID := uuid.MustParse("11111111-2222-3333-4444-555555555555")

	for _, tc := range []struct {
		name        string
		input       AddPageInput
		wantFormats []domain.Format
	}{
		{
			name:        "formats default to all",
			input:       AddPageInput{URL: "https://example.com"},
			wantFormats: domain.AllFormats(),
		},
		{
			name:        "requested formats are kept",
			input:       AddPageInput{URL: "https://example.com", Formats: []domain.Format{domain.FormatPDF}},
			wantFormats: []domain.Format{domain.FormatPDF},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := newFakeRepo()
			publisher := &fakePublisher{}
			archiver := &fakeArchiver{
				fetched: Fetched{
					Meta:   domain.Meta{Title: "Title"},
					Source: domain.Source{URL: tc.input.URL, Content: []byte("content")},
				},
			}

			service := NewService(repo, archiver, publisher, zaptest.NewLogger(t))
			service.newID = func() uuid.UUID { return fixedID }

			page, err := service.AddPage(context.Background(), tc.input)
			require.NoError(t, err)

			assert.Equal(t, fixedID, page.ID)
			assert.Equal(t, domain.StatusNew, page.Status)
			assert.Equal(t, "Title", page.Meta.Title)
			assert.Equal(t, tc.wantFormats, page.Formats)

			require.Len(t, repo.saved(), 1, "page must be stored before it is queued")
			require.Len(t, publisher.tasks, 1)
			assert.Equal(t, page, publisher.tasks[0].Page)
			assert.Equal(t, []byte("content"), publisher.tasks[0].Source.Content, "content travels with the task")
		})
	}
}

func TestService_AddPageValidation(t *testing.T) {
	t.Parallel()

	service := NewService(newFakeRepo(), &fakeArchiver{}, &fakePublisher{}, zaptest.NewLogger(t))

	_, err := service.AddPage(context.Background(), AddPageInput{})
	require.Error(t, err)

	var validation *ValidationError

	require.ErrorAs(t, err, &validation)
	assert.Equal(t, "url", validation.Field)
	assert.EqualError(t, validation.Err, "Value is required")
}

func TestService_AddPageFetchError(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	publisher := &fakePublisher{}
	archiver := &fakeArchiver{fetchErr: errors.New("dial tcp: timeout")}

	service := NewService(repo, archiver, publisher, zaptest.NewLogger(t))

	page, err := service.AddPage(context.Background(), AddPageInput{URL: "https://example.com"})
	require.NoError(t, err, "an unfetchable page is still registered")

	assert.Equal(t, "dial tcp: timeout", page.Meta.Error)
	assert.Equal(t, domain.StatusNew, page.Status)
	require.Len(t, publisher.tasks, 1)
}

func TestService_AddPagePushError(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()
	publisher := &fakePublisher{err: errors.New("queue is closed")}

	service := NewService(repo, &fakeArchiver{}, publisher, zaptest.NewLogger(t))

	_, err := service.AddPage(context.Background(), AddPageInput{URL: "https://example.com"})
	require.Error(t, err)

	assert.Len(t, repo.saved(), 1, "the page stays new and is picked up on the next start")
}

func TestService_GetPageNotFound(t *testing.T) {
	t.Parallel()

	service := NewService(newFakeRepo(), &fakeArchiver{}, &fakePublisher{}, zaptest.NewLogger(t))

	_, err := service.GetPage(context.Background(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_GetFileNotFound(t *testing.T) {
	t.Parallel()

	service := NewService(newFakeRepo(), &fakeArchiver{}, &fakePublisher{}, zaptest.NewLogger(t))

	_, err := service.GetFile(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestService_ProcessPage(t *testing.T) {
	t.Parallel()

	fileID := uuid.New()

	for _, tc := range []struct {
		name     string
		archiver *fakeArchiver
		formats  []domain.Format
		want     domain.Status
	}{
		{
			name:     "all formats archived",
			archiver: &fakeArchiver{files: []domain.File{domain.NewFile(fileID, "page.pdf", "application/pdf", nil, time.Now())}},
			formats:  []domain.Format{domain.FormatPDF, domain.FormatHeaders},
			want:     domain.StatusDone,
		},
		{
			name:     "one format failed",
			archiver: &fakeArchiver{archiveErr: errors.New("boom")},
			formats:  []domain.Format{domain.FormatPDF},
			want:     domain.StatusFailed,
		},
		{
			name:     "processor panicked",
			archiver: &fakeArchiver{panicWith: "boom"},
			formats:  []domain.Format{domain.FormatPDF},
			want:     domain.StatusFailed,
		},
		{
			name:     "no formats",
			archiver: &fakeArchiver{},
			formats:  nil,
			want:     domain.StatusFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := newFakeRepo()

			page := domain.NewPage(uuid.New(), "https://example.com", "", time.Now().UTC(), tc.formats...)

			service := NewService(repo, tc.archiver, &fakePublisher{}, zaptest.NewLogger(t))
			service.ProcessPage(context.Background(), ArchiveTask{
				Page:   page,
				Source: domain.Source{URL: page.URL, Content: []byte("content")},
			})

			assert.Equal(t, tc.want, page.Status)
			assert.Len(t, page.Results, len(tc.formats))
			assert.Len(t, tc.archiver.formats(), len(tc.formats), "every format must be tried")

			for _, source := range tc.archiver.archivedSource {
				assert.Equal(t, []byte("content"), source.Content, "formats reuse the fetched content")
			}

			require.Len(t, repo.saved(), 2, "processing state and results are stored separately")
			assert.Equal(t, domain.StatusProcessing, repo.saved()[0].Status)
			assert.Equal(t, tc.want, repo.saved()[1].Status)
		})
	}
}

func TestService_ProcessPageStoresError(t *testing.T) {
	t.Parallel()

	page := domain.NewPage(uuid.New(), "https://example.com", "", time.Now().UTC(), domain.FormatPDF)

	service := NewService(newFakeRepo(), &fakeArchiver{archiveErr: errors.New("dial tcp")}, &fakePublisher{}, zaptest.NewLogger(t))
	service.ProcessPage(context.Background(), ArchiveTask{Page: page})

	require.Len(t, page.Results, 1)
	assert.EqualError(t, page.Results[0].Err, "process: dial tcp")
}

func TestWorker_Run(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()

	pendingPage := domain.NewPage(uuid.New(), "https://example.com/pending", "", time.Now().UTC(), domain.FormatPDF)
	repo.pending = []*domain.Page{pendingPage}

	tasks := &fakeQueue{}
	service := NewService(repo, &fakeArchiver{}, &fakePublisher{}, zaptest.NewLogger(t))
	worker := NewWorker(service, repo, tasks, zaptest.NewLogger(t))

	tasks.closed = true

	require.NoError(t, worker.Run(context.Background()))

	stored, err := repo.Get(context.Background(), pendingPage.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDone, stored.Status, "pages left by a previous run are finished")
}

func TestWorker_RunProcessesQueuedTasks(t *testing.T) {
	t.Parallel()

	repo := newFakeRepo()

	queued := domain.NewPage(uuid.New(), "https://example.com/queued", "", time.Now().UTC(), domain.FormatPDF)

	tasks := &fakeQueue{tasks: []ArchiveTask{{Page: queued}}, closed: true}
	service := NewService(repo, &fakeArchiver{}, &fakePublisher{}, zaptest.NewLogger(t))
	worker := NewWorker(service, repo, tasks, zaptest.NewLogger(t))

	require.NoError(t, worker.Run(context.Background()))

	stored, err := repo.Get(context.Background(), queued.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.StatusDone, stored.Status)
}

type fakeQueue struct {
	tasks  []ArchiveTask
	closed bool
}

func (q *fakeQueue) Pop(ctx context.Context) (ArchiveTask, bool) {
	if len(q.tasks) > 0 {
		task := q.tasks[0]
		q.tasks = q.tasks[1:]

		return task, true
	}

	if q.closed {
		return ArchiveTask{}, false
	}

	<-ctx.Done()

	return ArchiveTask{}, false
}
