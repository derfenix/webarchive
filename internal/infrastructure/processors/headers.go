package processors

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/derfenix/webarchive/internal/domain"
)

func NewHeaders(client *http.Client) *Headers {
	return &Headers{client: client}
}

type Headers struct {
	client *http.Client
}

func (h *Headers) Process(ctx context.Context, source domain.Source) ([]domain.File, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, source.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call url: %w", err)
	}

	if resp.Body != nil {
		_ = resp.Body.Close()
	}

	file, err := h.newFile(resp.Header)
	if err != nil {
		return nil, fmt.Errorf("new file from headers: %w", err)
	}

	return []domain.File{file}, nil
}

func (h *Headers) newFile(headers http.Header) (domain.File, error) {
	buf := bytes.NewBuffer(nil)

	if err := headers.Write(buf); err != nil {
		return domain.File{}, fmt.Errorf("write headers: %w", err)
	}

	return newFile("headers", buf.Bytes()), nil
}
