package badger

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/derfenix/webarchive/adapters/repository"

	"github.com/derfenix/webarchive/entity"
)

func TestSite(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skip db test")
	}

	ctx := context.Background()

	tempDir, err := os.MkdirTemp(os.TempDir(), "badger_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		assert.NoError(t, os.RemoveAll(tempDir))
	})

	log := zaptest.NewLogger(t)

	db, err := repository.NewBadger(tempDir, log.Named("db"))
	require.NoError(t, err)

	siteRepo, err := NewPage(db)
	require.NoError(t, err)

	t.Run("base path", func(t *testing.T) {
		t.Parallel()

		site := entity.NewPage("https://google.com", "Save all google", entity.FormatPDF, entity.FormatSingleFile)
		site.Created = site.Created.Truncate(time.Microsecond)

		err := siteRepo.Save(ctx, site)
		require.NoError(t, err)

		storedSite, err := siteRepo.Get(ctx, site.ID)
		require.NoError(t, err)

		assert.Equal(t, site.ID, storedSite.ID)
		assert.Equal(t, site.URL, storedSite.URL)
		assert.Equal(t, site.Status, storedSite.Status)

		all, err := siteRepo.ListAll(ctx)
		require.NoError(t, err)
		require.Len(t, all, 1)

		assert.Equal(t, site.ID, all[0].ID)
		assert.Equal(t, site.URL, all[0].URL)
		assert.Equal(t, site.Status, all[0].Status)
	})
}
