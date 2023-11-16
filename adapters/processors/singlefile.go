package processors

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/url"

	"golang.org/x/net/html"

	"github.com/derfenix/webarchive/entity"
)

func NewSingleFile(client *http.Client) *SingleFile {
	return &SingleFile{client: client}
}

type SingleFile struct {
	client *http.Client
}

func (s *SingleFile) Process(ctx context.Context, url string, cache *entity.Cache) ([]entity.File, error) {
	reader := cache.Reader()

	if reader == nil {
		response, err := s.get(ctx, url)
		if err != nil {
			return nil, err
		}

		if response.Body != nil {
			defer func() {
				_ = response.Body.Close()
			}()
		}

		reader = response.Body
	}

	htmlNode, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parse response body: %w", err)
	}

	if err := s.process(ctx, htmlNode, url); err != nil {
		return nil, fmt.Errorf("process: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := html.Render(buf, htmlNode); err != nil {
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

func (s *SingleFile) process(ctx context.Context, node *html.Node, pageURL string) error {
	parsedURL, err := url.Parse(pageURL)
	if err != nil {
		return fmt.Errorf("parse page url: %w", err)
	}

	baseURL := fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		var err error
		switch child.Data {
		case "head":
			err = s.processHead(ctx, child, baseURL)

		case "body":
			err = s.processBody(ctx, child, baseURL)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

func (s *SingleFile) processHead(ctx context.Context, node *html.Node, baseURL string) error {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		switch child.Data {
		case "link":
			if err := s.processHref(ctx, child.Attr, baseURL); err != nil {
				return fmt.Errorf("process link %s: %w", child.Attr, err)
			}

		case "script":
			if err := s.processSrc(ctx, child.Attr, baseURL); err != nil {
				return fmt.Errorf("process script %s: %w", child.Attr, err)
			}
		}
	}

	return nil
}

func (s *SingleFile) processBody(ctx context.Context, child *html.Node, url string) error {
	return nil
}

func (s *SingleFile) processHref(ctx context.Context, attrs []html.Attribute, baseURL string) error {
	return nil
}

func (s *SingleFile) processSrc(ctx context.Context, attrs []html.Attribute, baseURL string) error {
	return nil
}
