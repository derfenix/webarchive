package processors

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/net/html"

	"github.com/derfenix/webarchive/adapters/processors/internal"
	"github.com/derfenix/webarchive/entity"
)

func NewSingleFile(client *http.Client, log *zap.Logger) *SingleFile {
	return &SingleFile{client: client, log: log}
}

type SingleFile struct {
	client *http.Client
	log    *zap.Logger
}

func (s *SingleFile) Process(ctx context.Context, pageURL string, cache *entity.Cache) ([]entity.File, error) {
	reader := cache.Reader()

	if reader == nil {
		response, err := s.get(ctx, pageURL)
		if err != nil {
			return nil, err
		}

		defer func() {
			_ = response.Body.Close()
		}()

		reader = response.Body
	}

	inlinedHTML, err := internal.NewMediaInline(s.log, s.get).Inline(ctx, reader, pageURL)
	if err != nil {
		return nil, fmt.Errorf("inline media: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := html.Render(buf, inlinedHTML); err != nil {
		return nil, fmt.Errorf("render result html: %w", err)
	}

	htmlFile := entity.NewFile("page.html", buf.Bytes())

	return []entity.File{htmlFile}, nil
}

func (s *SingleFile) get(ctx context.Context, url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("new request: %w", err)
	}

	response, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("want status 200, got %d", response.StatusCode)
	}

	if response.Body == nil {
		return nil, fmt.Errorf("empty response body")
	}

	return response, nil
}
