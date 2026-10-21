// Package rest is the transport adapter of the HTTP API: it maps generated
// ogen types to the use cases and back, and holds no archive rules of its own.
package rest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/derfenix/webarchive/internal/api/openapi"
	"github.com/derfenix/webarchive/internal/application"
	"github.com/derfenix/webarchive/internal/domain"
)

// Pages is the transport's port to the archive use cases.
type Pages interface {
	AddPage(ctx context.Context, in application.AddPageInput) (*domain.Page, error)
	GetPage(ctx context.Context, id uuid.UUID) (*domain.Page, error)
	ListPages(ctx context.Context) ([]*domain.Page, error)
	GetFile(ctx context.Context, pageID, fileID uuid.UUID) (*domain.File, error)
}

func NewHandler(pages Pages) *Handler {
	return &Handler{pages: pages}
}

type Handler struct {
	openapi.UnimplementedHandler
	pages Pages
}

func (h *Handler) GetPage(ctx context.Context, params openapi.GetPageParams) (openapi.GetPageRes, error) {
	page, err := h.pages.GetPage(ctx, params.ID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &openapi.GetPageNotFound{}, nil
		}

		return nil, err
	}

	restPage := PageToRestWithResults(page)

	return &restPage, nil
}

func (h *Handler) AddPage(ctx context.Context, req openapi.OptAddPageReq, params openapi.AddPageParams) (openapi.AddPageRes, error) {
	url := params.URL.Or(req.Value.URL)
	description := params.Description.Or(req.Value.Description.Value)

	formats := req.Value.Formats
	if len(formats) == 0 {
		formats = params.Formats
	}
	if len(formats) == 0 {
		formats = []openapi.Format{openapi.FormatAll}
	}

	switch {
	case req.Value.URL != "":
		url = req.Value.URL
	case params.URL.IsSet():
		url = params.URL.Value
	}

	domainFormats, err := FormatFromRest(formats)
	if err != nil {
		return &openapi.AddPageBadRequest{
			Field: "formats",
			Error: err.Error(),
		}, nil
	}

	page, err := h.pages.AddPage(ctx, application.AddPageInput{
		URL:         url,
		Description: description,
		Formats:     domainFormats,
	})
	if err != nil {
		var validation *application.ValidationError
		if errors.As(err, &validation) {
			return &openapi.AddPageBadRequest{
				Field: validation.Field,
				Error: validation.Err.Error(),
			}, nil
		}

		return nil, err
	}

	res := BasePageToRest(&page.PageBase)

	return &res, nil
}

func (h *Handler) GetPages(ctx context.Context) (openapi.Pages, error) {
	pages, err := h.pages.ListPages(ctx)
	if err != nil {
		return nil, err
	}

	res := make(openapi.Pages, len(pages))
	for i := range res {
		res[i] = PageToRest(pages[i])
	}

	return res, nil
}

func (h *Handler) GetFile(ctx context.Context, params openapi.GetFileParams) (openapi.GetFileRes, error) {
	file, err := h.pages.GetFile(ctx, params.ID, params.FileID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return &openapi.GetFileNotFound{}, nil
		}

		return nil, err
	}

	switch {
	case file.MimeType == "application/pdf":
		return &openapi.GetFileOKApplicationPdf{Data: bytes.NewReader(file.Data)}, nil

	case strings.HasPrefix(file.MimeType, "text/plain"):
		return &openapi.GetFileOKTextPlain{Data: bytes.NewReader(file.Data)}, nil

	case strings.HasPrefix(file.MimeType, "text/html"):
		return &openapi.GetFileOKTextHTML{Data: bytes.NewReader(file.Data)}, nil

	default:
		return nil, fmt.Errorf("unsupported mimetype: %s", file.MimeType)
	}
}

func (h *Handler) NewError(_ context.Context, err error) *openapi.ErrorStatusCode {
	return &openapi.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: openapi.Error{
			Message:   err.Error(),
			Localized: openapi.OptString{},
		},
	}
}
