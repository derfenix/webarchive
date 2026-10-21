package domain

import (
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPage(t *testing.T) *Page {
	t.Helper()

	return NewPage(uuid.New(), "https://example.com", "example", time.Now().UTC(), FormatPDF)
}

func TestNewPage(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	created := time.Now().UTC()

	page := NewPage(id, "https://example.com", "example", created, FormatPDF, FormatHeaders)

	assert.Equal(t, id, page.ID)
	assert.Equal(t, "https://example.com", page.URL)
	assert.Equal(t, "example", page.Description)
	assert.Equal(t, created, page.Created)
	assert.Equal(t, []Format{FormatPDF, FormatHeaders}, page.Formats)
	assert.Equal(t, StatusNew, page.Status)
	assert.Equal(t, uint16(1), page.Version)
}

func TestPage_SetMetaError(t *testing.T) {
	t.Parallel()

	page := newPage(t)
	page.SetMetaError(errors.New("boom"))

	assert.Equal(t, "boom", page.Meta.Error)
}

func TestPage_Finish(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name    string
		results []Result
		want    Status
	}{
		{
			name:    "all formats archived",
			results: []Result{{Format: FormatPDF}, {Format: FormatHeaders}},
			want:    StatusDone,
		},
		{
			name:    "some formats failed",
			results: []Result{{Format: FormatPDF}, {Format: FormatHeaders, Err: errors.New("boom")}},
			want:    StatusWithErrors,
		},
		{
			name:    "every format failed",
			results: []Result{{Format: FormatPDF, Err: errors.New("boom")}},
			want:    StatusFailed,
		},
		{
			name:    "no formats at all",
			results: nil,
			want:    StatusFailed,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			page := newPage(t)
			page.StartProcessing()
			page.Finish(tc.results)

			assert.Equal(t, tc.want, page.Status)
			assert.Equal(t, ResultsRO(tc.results), page.Results)
		})
	}
}

func TestAllFormats(t *testing.T) {
	t.Parallel()

	formats := AllFormats()
	assert.Equal(t, []Format{FormatHeaders, FormatPDF, FormatSingleFile}, formats)

	formats[0] = FormatPDF
	assert.Equal(t, []Format{FormatHeaders, FormatPDF, FormatSingleFile}, AllFormats(), "callers must not share the slice")
}

func TestNewFile(t *testing.T) {
	t.Parallel()

	id := uuid.New()
	created := time.Now().UTC()

	file := NewFile(id, "page.html", "text/html", []byte("hello"), created)

	assert.Equal(t, id, file.ID)
	assert.Equal(t, "page.html", file.Name)
	assert.Equal(t, "text/html", file.MimeType)
	assert.Equal(t, int64(5), file.Size)
	assert.Equal(t, created, file.Created)
}

func TestSource(t *testing.T) {
	t.Parallel()

	empty := Source{URL: "https://example.com"}
	assert.False(t, empty.HasContent())

	source := Source{URL: "https://example.com", Content: []byte("hello")}
	assert.True(t, source.HasContent())

	content, err := io.ReadAll(source.Reader())
	require.NoError(t, err)
	assert.Equal(t, "hello", string(content))
}
