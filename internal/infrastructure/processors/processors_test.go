package processors

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"

	"github.com/derfenix/webarchive/internal/config"
	"github.com/derfenix/webarchive/internal/domain"
)

const testPage = `<!DOCTYPE html>
<html><head><title>Test page</title><meta name="description" content="Just a test"></head>
<body><p>hello</p></body></html>`

func newTestProcessors(t *testing.T) *Processors {
	t.Helper()

	procs, err := NewProcessors(config.PDF{}, zaptest.NewLogger(t))
	require.NoError(t, err)

	return procs
}

func TestProcessors_Fetch(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(testPage))
	}))
	t.Cleanup(server.Close)

	fetched, err := newTestProcessors(t).Fetch(ctx, server.URL)
	require.NoError(t, err)

	assert.Equal(t, "Test page", fetched.Meta.Title)
	assert.Equal(t, "Just a test", fetched.Meta.Description)
	assert.Equal(t, "utf-8", fetched.Meta.Encoding)
	assert.Equal(t, server.URL, fetched.Source.URL)
	assert.Equal(t, testPage, string(fetched.Source.Content), "content is reused by the formats")
}

func TestProcessors_FetchBadStatus(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	_, err := newTestProcessors(t).Fetch(context.Background(), server.URL)
	require.Error(t, err)
}

func TestProcessors_ArchiveHeaders(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("X-Test", "yes")
	}))
	t.Cleanup(server.Close)

	files, err := newTestProcessors(t).Archive(ctx, domain.FormatHeaders, domain.Source{URL: server.URL})
	require.NoError(t, err)
	require.Len(t, files, 1)

	assert.Contains(t, string(files[0].Data), "X-Test: yes")
	assert.Contains(t, files[0].MimeType, "text/plain")
	assert.Equal(t, int64(len(files[0].Data)), files[0].Size)
}

func TestProcessors_ArchiveUnknownFormat(t *testing.T) {
	t.Parallel()

	_, err := newTestProcessors(t).Archive(context.Background(), domain.Format(99), domain.Source{})
	require.Error(t, err)
}
