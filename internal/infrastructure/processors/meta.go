package processors

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"

	"golang.org/x/net/html"

	"github.com/derfenix/webarchive/internal/application"
	"github.com/derfenix/webarchive/internal/domain"
)

const defaultEncoding = "utf-8"

// Fetch requests the page once and returns its meta together with the content,
// so the formats do not download the page again.
func (p *Processors) Fetch(ctx context.Context, pageURL string) (application.Fetched, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
	if err != nil {
		return application.Fetched{}, fmt.Errorf("new request: %w", err)
	}

	response, err := p.client.Do(req)
	if err != nil {
		return application.Fetched{}, fmt.Errorf("do request: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return application.Fetched{}, fmt.Errorf("want status 200, got %d", response.StatusCode)
	}

	if response.Body == nil {
		return application.Fetched{}, fmt.Errorf("empty response body")
	}

	defer func() {
		_ = response.Body.Close()
	}()

	content := bytes.NewBuffer(nil)

	htmlNode, err := html.Parse(io.TeeReader(response.Body, content))
	if err != nil {
		return application.Fetched{}, fmt.Errorf("parse response body: %w", err)
	}

	head, err := findHead(htmlNode)
	if err != nil {
		return application.Fetched{}, err
	}

	meta := domain.Meta{}
	getMetaData(head, &meta)
	meta.Encoding = encodingFromHeader(response.Header)

	return application.Fetched{
		Meta:   meta,
		Source: domain.Source{URL: pageURL, Content: content.Bytes()},
	}, nil
}

// findHead looks for the head element of the document. Walking for the element
// itself keeps working for pages without a doctype, which the previous
// first-child scan did not handle.
func findHead(root *html.Node) (*html.Node, error) {
	var found *html.Node

	var walk func(n *html.Node)

	walk = func(n *html.Node) {
		if n == nil || found != nil {
			return
		}

		if n.Type == html.ElementNode && n.Data == "head" {
			found = n

			return
		}

		for child := n.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(root)

	if found == nil {
		return nil, fmt.Errorf("failed to find head tag")
	}

	return found, nil
}

func getMetaData(n *html.Node, meta *domain.Meta) {
	if n == nil {
		return
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.ElementNode && c.Data == "title" && c.FirstChild != nil {
			meta.Title = c.FirstChild.Data
		}

		if c.Type == html.ElementNode && c.Data == "meta" {
			attrs := make(map[string]string)
			for _, attr := range c.Attr {
				attrs[attr.Key] = attr.Val
			}

			name, ok := attrs["name"]
			if ok && name == "description" {
				meta.Description = attrs["content"]
			}
		}

		getMetaData(c, meta)
	}
}

func encodingFromHeader(headers http.Header) string {
	var foundEncoding bool
	var encoding string

	_, encoding, foundEncoding = strings.Cut(headers.Get("Content-Type"), "; ")
	if foundEncoding {
		_, encoding, foundEncoding = strings.Cut(encoding, "=")
	}

	if !foundEncoding {
		encoding = defaultEncoding
	}

	return encoding
}
