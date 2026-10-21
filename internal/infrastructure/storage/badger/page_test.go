package badger

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/derfenix/webarchive/internal/domain"
	"github.com/derfenix/webarchive/internal/infrastructure/storage"
)

// openRepo gives a subtest its own database, so subtests can run in parallel.
func openRepo(t *testing.T) *Page {
	t.Helper()

	tempDir, err := os.MkdirTemp(os.TempDir(), "badger_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	})

	log := zaptest.NewLogger(t)

	db, err := storage.Open(tempDir, log.Named("db"))
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, db.Close())
	})

	repo, err := NewPage(db)
	require.NoError(t, err)

	return repo
}

func TestPage_SaveGetList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := openRepo(t)

	page := domain.NewPage(uuid.New(), "https://google.com", "Save all google", time.Now().UTC(),
		domain.FormatPDF, domain.FormatSingleFile)
	page.Created = page.Created.Truncate(time.Microsecond)

	require.NoError(t, repo.Save(ctx, page))

	stored, err := repo.Get(ctx, page.ID)
	require.NoError(t, err)

	assert.Equal(t, page.ID, stored.ID)
	assert.Equal(t, page.URL, stored.URL)
	assert.Equal(t, page.Status, stored.Status)

	all, err := repo.List(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)

	assert.Equal(t, page.ID, all[0].ID)
	assert.Equal(t, page.URL, all[0].URL)
	assert.Equal(t, page.Status, all[0].Status)
}

func TestPage_Missing(t *testing.T) {
	t.Parallel()

	_, err := openRepo(t).Get(context.Background(), uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

func TestPage_ListPending(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := openRepo(t)

	pending := domain.NewPage(uuid.New(), "https://example.com/pending", "", time.Now().UTC(), domain.FormatPDF)
	require.NoError(t, repo.Save(ctx, pending))

	processing := domain.NewPage(uuid.New(), "https://example.com/processing", "", time.Now().UTC(), domain.FormatPDF)
	processing.StartProcessing()
	require.NoError(t, repo.Save(ctx, processing))

	done := domain.NewPage(uuid.New(), "https://example.com/done", "", time.Now().UTC(), domain.FormatPDF)
	done.Status = domain.StatusDone
	require.NoError(t, repo.Save(ctx, done))

	list, err := repo.ListPending(ctx)
	require.NoError(t, err)

	ids := make([]uuid.UUID, 0, len(list))
	for i := range list {
		ids = append(ids, list[i].ID)
	}

	assert.Contains(t, ids, pending.ID)
	assert.Contains(t, ids, processing.ID, "an interrupted page is picked up again")
	assert.NotContains(t, ids, done.ID)
}

func TestPage_GetFile(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	repo := openRepo(t)

	page := fixturePage()
	require.NoError(t, repo.Save(ctx, page))

	file, err := repo.GetFile(ctx, page.ID, fixtureFileID)
	require.NoError(t, err)
	assert.Equal(t, fixtureFileID, file.ID)
	assert.Equal(t, "application/pdf", file.MimeType)

	_, err = repo.GetFile(ctx, page.ID, uuid.New())
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)

	_, err = repo.GetFile(ctx, uuid.New(), fixtureFileID)
	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
