package rest

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/derfenix/webarchive/api/openapi"
	"github.com/derfenix/webarchive/entity"
)

type Pages interface {
	ListAll(ctx context.Context) ([]*entity.Page, error)
	Save(ctx context.Context, site *entity.Page) error
	Get(_ context.Context, id uuid.UUID) (*entity.Page, error)
}

func NewService(sites Pages) *Service {
	return &Service{pages: sites}
}

type Service struct {
	openapi.UnimplementedHandler
	pages Pages
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
	site := entity.NewPage(req.Value.URL, req.Value.Description.Value, FormatFromRest(req.Value.Formats)...)

	err := s.pages.Save(ctx, site)
	if err != nil {
		return nil, fmt.Errorf("save site: %w", err)
	}

	res := PageToRest(site)

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

func (s *Service) NewError(_ context.Context, err error) *openapi.ErrorStatusCode {
	return &openapi.ErrorStatusCode{
		StatusCode: http.StatusInternalServerError,
		Response: openapi.Error{
			Message:   err.Error(),
			Localized: openapi.OptString{},
		},
	}
}
