package queue

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/derfenix/webarchive/internal/application"
	"github.com/derfenix/webarchive/internal/domain"
)

func testTask() application.ArchiveTask {
	return application.ArchiveTask{
		Page:   domain.NewPage(uuid.New(), "https://example.com", "", time.Now().UTC(), domain.FormatPDF),
		Source: domain.Source{URL: "https://example.com", Content: []byte("content")},
	}
}

func TestMemory_PushPop(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	q := NewMemory(1)

	task := testTask()

	require.NoError(t, q.Push(ctx, task))

	got, open := q.Pop(ctx)
	require.True(t, open)
	assert.Equal(t, task, got)
}

func TestMemory_ClosedQueue(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	q := NewMemory(1)

	require.NoError(t, q.Close())

	_, open := q.Pop(ctx)
	assert.False(t, open, "a closed queue stops the worker")
}

func TestMemory_CancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	assert.Error(t, NewMemory(0).Push(ctx, testTask()), "push does not hang when the caller is done")

	_, open := NewMemory(0).Pop(ctx)
	assert.False(t, open)
}
