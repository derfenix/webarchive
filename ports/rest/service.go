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

func NewService(pages Pages, ch chan *entity.Page, processor entity.Processor) *Service {
	return &Service{
		pages:     pages,
		ch:        ch,
		processor: processor,
	}
}

type Service struct {
	openapi.UnimplementedHandler
	processor entity.Processor
	pages     Pages
	ch        chan *entity.Page
}

func (s *Service) GetPage(ctx context.Context, params openapi.GetPageParams) (openapi.GetPageRes, error) {
	page, err := s.pages.Get(ctx, params.ID)
	if err != nil {
		return &openapi.GetPageNotFound{}, nil
	}

	restPage := PageToRestWithResults(page)

	return &restPage, nil
}

func (s *Service) AddPage(ctx context.Context, req openapi.OptAddPageReq, params openapi.AddPageParams) (openapi.AddPageRes, error) {
	url := params.URL.Or(req.Value.URL)
	description := params.Description.Or(req.Value.Description.Value)

	formats := req.Value.Formats
	if len(formats) == 0 {
		formats = params.Formats
	}
	if len(formats) == 0 {
		formats = []openapi.Format{"all"}
	}

	switch {
	case req.Value.URL != "":
		url = req.Value.URL
	case params.URL.IsSet():
		url = params.URL.Value
	}

	if url == "" {
		return &openapi.AddPageBadRequest{
			Field: "url",
			Error: "Value is required",
		}, nil
	}

	domainFormats, err := FormatFromRest(formats)
	if err != nil {
		return &openapi.AddPageBadRequest{
			Field: "formats",
			Error: err.Error(),
		}, nil
	}

	page := entity.NewPage(url, description, domainFormats...)
	page.Status = entity.StatusNew
	page.Prepare(ctx, s.processor)

	if err := s.pages.Save(ctx, page); err != nil {
		return nil, fmt.Errorf("save page: %w", err)
	}

	res := BasePageToRest(&page.PageBase)

	s.ch <- page

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
