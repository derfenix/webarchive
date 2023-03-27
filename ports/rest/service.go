package rest

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/derfenix/webarchive/api/openapi"
	"github.com/derfenix/webarchive/entity"
)

type Pages interface {
	ListAll(ctx context.Context) ([]*entity.Page, error)
	Save(ctx context.Context, site *entity.Page) error
	Get(ctx context.Context, id uuid.UUID) (*entity.Page, error)
	GetFile(ctx context.Context, pageID, fileID uuid.UUID) (*entity.File, error)
}

func NewService(sites Pages, ch chan *entity.Page) *Service {
	return &Service{pages: sites, ch: ch}
}

type Service struct {
	openapi.UnimplementedHandler
	pages Pages
	ch    chan *entity.Page
}

func (s *Service) GetPage(ctx context.Context, params openapi.GetPageParams) (openapi.GetPageRes, error) {
	page, err := s.pages.Get(ctx, params.ID)
	if err != nil {
		return &openapi.GetPageNotFound{}, nil
	}

	restPage := PageToRestWithResults(page)

	return &restPage, nil
}

func (s *Service) AddPage(ctx context.Context, req openapi.OptAddPageReq) (*openapi.Page, error) {
	page := entity.NewPage(req.Value.URL, req.Value.Description.Value, FormatFromRest(req.Value.Formats)...)

	page.Status = entity.StatusProcessing

	err := s.pages.Save(ctx, page)
	if err != nil {
		return nil, fmt.Errorf("save page: %w", err)
	}

	s.ch <- page

	res := PageToRest(page)

	return &res, nil
}

func (s *Service) GetPages(ctx context.Context) (openapi.Pages, error) {
	sites, err := s.pages.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("list all: %w", err)
	}

	res := make(openapi.Pages, len(sites))
	for i := range res {
		res[i] = PageToRest(sites[i])
	}

	return res, nil
}

func (s *Service) GetFile(ctx context.Context, params openapi.GetFileParams) (openapi.GetFileRes, error) {
	file, err := s.pages.GetFile(ctx, params.ID, params.FileID)
	if err != nil {
		return &openapi.GetFileNotFound{}, nil
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

func (s *Service) NewError(_ context.Context, err error) *openapi.ErrorStatusCode {
	return &openapi.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: openapi.Error{
			Message:   err.Error(),
			Localized: openapi.OptString{},
		},
	}
}
