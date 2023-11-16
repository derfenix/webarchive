package processors

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/derfenix/webarchive/config"
)

func TestProcessors_GetMeta(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg, err := config.NewConfig(ctx)
	require.NoError(t, err)

	procs, err := NewProcessors(cfg)
	require.NoError(t, err)

	meta, err := procs.GetMeta(ctx, "https://habr.com/ru/companies/wirenboard/articles/722718/")
	require.NoError(t, err)
	assert.Equal(t, "Сколько стоит умный дом? Рассказываю, как строил свой и что получилось за 1000 руб./м² / Хабр", meta.Title)
}
