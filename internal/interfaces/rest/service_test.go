package rest

import (
	"context"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/derfenix/webarchive/internal/api/openapi"
	"github.com/derfenix/webarchive/internal/application"
	"github.com/derfenix/webarchive/internal/domain"
)

type fakePages struct {
	page     *domain.Page
	pageErr  error
	list     []*domain.Page
	listErr  error
	file     *domain.File
	fileErr  error
	addErr   error
	addInput application.AddPageInput
}

func (p *fakePages) AddPage(_ context.Context, in application.AddPageInput) (*domain.Page, error) {
	p.addInput = in

	if p.addErr != nil {
		return nil, p.addErr
	}

	return p.page, nil
}

func (p *fakePages) GetPage(context.Context, uuid.UUID) (*domain.Page, error) {
	return p.page, p.pageErr
}

func (p *fakePages) ListPages(context.Context) ([]*domain.Page, error) {
	return p.list, p.listErr
}

func (p *fakePages) GetFile(context.Context, uuid.UUID, uuid.UUID) (*domain.File, error) {
	return p.file, p.fileErr
}

func testPage() *domain.Page {
	page := domain.NewPage(uuid.New(), "https://example.com", "example", time.Now().UTC(), domain.FormatPDF)
	page.Status = domain.StatusDone
	page.Meta = domain.Meta{Title: "Title", Description: "Description"}

	return page
}

func TestHandler_GetPage(t *testing.T) {
	t.Parallel()

	page := testPage()

	res, err := NewHandler(&fakePages{page: page}).GetPage(context.Background(), openapi.GetPageParams{ID: page.ID})
	require.NoError(t, err)

	ok, isOK := res.(*openapi.PageWithResults)
	require.True(t, isOK, "want the page, got %T", res)

	assert.Equal(t, page.ID, ok.ID)
	assert.Equal(t, page.URL, ok.URL)
	assert.Equal(t, openapi.StatusDone, ok.Status)
	assert.Empty(t, ok.Results)
}

func TestHandler_GetPageNotFound(t *testing.T) {
	t.Parallel()

	res, err := NewHandler(&fakePages{pageErr: domain.ErrNotFound}).
		GetPage(context.Background(), openapi.GetPageParams{ID: uuid.New()})
	require.NoError(t, err)

	assert.IsType(t, &openapi.GetPageNotFound{}, res)
}

func TestHandler_GetPageStorageError(t *testing.T) {
	t.Parallel()

	_, err := NewHandler(&fakePages{pageErr: errors.New("db is closed")}).
		GetPage(context.Background(), openapi.GetPageParams{ID: uuid.New()})
	require.Error(t, err, "an internal failure is not a 404")
}

func TestHandler_AddPage(t *testing.T) {
	t.Parallel()

	page := testPage()
	pages := &fakePages{page: page}

	req := openapi.OptAddPageReq{
		Set: true,
		Value: openapi.AddPageReq{
			URL:         "https://example.com",
			Description: openapi.NewOptString("described"),
			Formats:     []openapi.Format{openapi.FormatPdf},
		},
	}

	res, err := NewHandler(pages).AddPage(context.Background(), req, openapi.AddPageParams{})
	require.NoError(t, err)

	ok, isOK := res.(*openapi.Page)
	require.True(t, isOK, "want the page, got %T", res)

	assert.Equal(t, page.ID, ok.ID)
	assert.Equal(t, "https://example.com", pages.addInput.URL)
	assert.Equal(t, "described", pages.addInput.Description)
	assert.Equal(t, []domain.Format{domain.FormatPDF}, pages.addInput.Formats)
}

func TestHandler_AddPageFromQuery(t *testing.T) {
	t.Parallel()

	pages := &fakePages{page: testPage()}

	params := openapi.AddPageParams{
		URL:         openapi.NewOptString("https://example.com/query"),
		Description: openapi.NewOptString("from query"),
		Formats:     []openapi.Format{openapi.FormatHeaders},
	}

	_, err := NewHandler(pages).AddPage(context.Background(), openapi.OptAddPageReq{}, params)
	require.NoError(t, err)

	assert.Equal(t, "https://example.com/query", pages.addInput.URL)
	assert.Equal(t, "from query", pages.addInput.Description)
	assert.Equal(t, []domain.Format{domain.FormatHeaders}, pages.addInput.Formats)
}

func TestHandler_AddPageAllFormats(t *testing.T) {
	t.Parallel()

	pages := &fakePages{page: testPage()}

	_, err := NewHandler(pages).AddPage(context.Background(), openapi.OptAddPageReq{
		Set:   true,
		Value: openapi.AddPageReq{URL: "https://example.com"},
	}, openapi.AddPageParams{})
	require.NoError(t, err)

	assert.Equal(t, domain.AllFormats(), pages.addInput.Formats)
}

func TestHandler_AddPageInvalidFormat(t *testing.T) {
	t.Parallel()

	res, err := NewHandler(&fakePages{}).AddPage(context.Background(), openapi.OptAddPageReq{
		Set: true,
		Value: openapi.AddPageReq{
			URL:     "https://example.com",
			Formats: []openapi.Format{"markdown"},
		},
	}, openapi.AddPageParams{})
	require.NoError(t, err)

	badRequest, isBadRequest := res.(*openapi.AddPageBadRequest)
	require.True(t, isBadRequest, "want a bad request, got %T", res)

	assert.Equal(t, "formats", badRequest.Field)
}

func TestHandler_AddPageValidationError(t *testing.T) {
	t.Parallel()

	res, err := NewHandler(&fakePages{
		addErr: &application.ValidationError{Field: "url", Err: errors.New("Value is required")},
	}).AddPage(context.Background(), openapi.OptAddPageReq{}, openapi.AddPageParams{})
	require.NoError(t, err)

	badRequest, isBadRequest := res.(*openapi.AddPageBadRequest)
	require.True(t, isBadRequest, "want a bad request, got %T", res)

	assert.Equal(t, "url", badRequest.Field)
	assert.Equal(t, "Value is required", badRequest.Error)
}

func TestHandler_GetPages(t *testing.T) {
	t.Parallel()

	page := testPage()

	pages, err := NewHandler(&fakePages{list: []*domain.Page{page}}).GetPages(context.Background())
	require.NoError(t, err)

	require.Len(t, pages, 1)
	assert.Equal(t, page.ID, pages[0].ID)
}

func TestHandler_GetFile(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		mimeType string
		want     any
	}{
		{name: "pdf", mimeType: "application/pdf", want: &openapi.GetFileOKApplicationPdf{}},
		{name: "html", mimeType: "text/html; charset=utf-8", want: &openapi.GetFileOKTextHTML{}},
		{name: "plain", mimeType: "text/plain; charset=utf-8", want: &openapi.GetFileOKTextPlain{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			file := domain.NewFile(uuid.New(), "file", tc.mimeType, []byte("content"), time.Now())

			res, err := NewHandler(&fakePages{file: &file}).
				GetFile(context.Background(), openapi.GetFileParams{ID: uuid.New(), FileID: file.ID})
			require.NoError(t, err)

			assert.IsType(t, tc.want, res)
		})
	}
}

func TestHandler_GetFileContent(t *testing.T) {
	t.Parallel()

	file := domain.NewFile(uuid.New(), "page.pdf", "application/pdf", []byte("%PDF-1.4"), time.Now())

	res, err := NewHandler(&fakePages{file: &file}).
		GetFile(context.Background(), openapi.GetFileParams{ID: uuid.New(), FileID: file.ID})
	require.NoError(t, err)

	pdf, isPDF := res.(*openapi.GetFileOKApplicationPdf)
	require.True(t, isPDF, "want a pdf, got %T", res)

	data, err := io.ReadAll(pdf.Data)
	require.NoError(t, err)
	assert.Equal(t, "%PDF-1.4", string(data))
}

func TestHandler_GetFileNotFound(t *testing.T) {
	t.Parallel()

	res, err := NewHandler(&fakePages{fileErr: domain.ErrNotFound}).
		GetFile(context.Background(), openapi.GetFileParams{ID: uuid.New(), FileID: uuid.New()})
	require.NoError(t, err)

	assert.IsType(t, &openapi.GetFileNotFound{}, res)
}

func TestHandler_GetFileUnsupportedMimeType(t *testing.T) {
	t.Parallel()

	file := domain.NewFile(uuid.New(), "file", "application/zip", nil, time.Now())

	_, err := NewHandler(&fakePages{file: &file}).
		GetFile(context.Background(), openapi.GetFileParams{ID: uuid.New(), FileID: file.ID})
	require.Error(t, err)
}

func TestHandler_NewError(t *testing.T) {
	t.Parallel()

	res := NewHandler(&fakePages{}).NewError(context.Background(), errors.New("boom"))

	assert.Equal(t, http.StatusInternalServerError, res.StatusCode)
	assert.Equal(t, "boom", res.Response.Message)
}
