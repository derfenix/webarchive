package processors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/derfenix/webarchive/config"
	"github.com/derfenix/webarchive/entity"
)

func TestProcessors_GetMeta(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg, err := config.NewConfig(ctx)
	require.NoError(t, err)

	procs, err := NewProcessors(cfg, zaptest.NewLogger(t))
	require.NoError(t, err)

	cache := entity.NewCache()

	meta, err := procs.GetMeta(ctx, "https://habr.com/ru/companies/wirenboard/articles/722718/", cache)
	require.NoError(t, err)
	assert.Equal(t, "Сколько стоит умный дом? Рассказываю, как строил свой и что получилось за 1000 руб./м² / Хабр", meta.Title)
}
