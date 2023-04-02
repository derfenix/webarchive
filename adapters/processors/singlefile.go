package processors

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"

	"github.com/derfenix/webarchive/entity"
)

const defaultEncoding = "utf-8"

func NewSingleFile(client *http.Client) *SingleFile {
	return &SingleFile{client: client}
}

type SingleFile struct {
	client *http.Client
}

func (s *SingleFile) Process(ctx context.Context, url string) ([]entity.File, error) {
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

	defer func() {
		_ = response.Body.Close()
	}()

	htmlNode, err := html.Parse(response.Body)
	if err != nil {
		return nil, fmt.Errorf("parse response body: %w", err)
	}

	if err := s.crawl(ctx, htmlNode, baseURL(url), getEncoding(response)); err != nil {
		return nil, fmt.Errorf("crawl: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := html.Render(buf, htmlNode); err != nil {
		return nil, fmt.Errorf("render result html: %w", err)
	}

	htmlFile := entity.NewFile("page.html", buf.Bytes())

	return []entity.File{htmlFile}, nil
}

func (s *SingleFile) crawl(ctx context.Context, node *html.Node, baseURL string, encoding string) error {
	if node.Data == "head" {
		s.setCharset(node, encoding)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Type == html.ElementNode {
			if err := s.findAndReplaceResources(ctx, child, baseURL); err != nil {
				return err
			}
		}

		if err := s.crawl(ctx, child, baseURL, encoding); err != nil {
			return fmt.Errorf("crawl child %s: %w", child.Data, err)
		}
	}

	return nil
}

func (s *SingleFile) findAndReplaceResources(ctx context.Context, node *html.Node, baseURL string) error {
	switch node.DataAtom {
	case atom.Img, atom.Image, atom.Script, atom.Style:
		err := s.replaceResource(ctx, node, baseURL)
		if err != nil {
			return err
		}

	case atom.Link:
		for _, attribute := range node.Attr {
			if attribute.Key == "rel" && (attribute.Val == "stylesheet") {
				if err := s.replaceResource(ctx, node, baseURL); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (s *SingleFile) replaceResource(ctx context.Context, node *html.Node, baseURL string) error {
	for i, attribute := range node.Attr {
		if attribute.Key == "src" || attribute.Key == "href" {
			encoded, contentType, err := s.loadResource(ctx, attribute.Val, baseURL)
			if err != nil {
				return fmt.Errorf("load resource for %s: %w", node.Data, err)
			}

			if len(encoded) == 0 {
				attribute.Val = ""

			} else {
				attribute.Val = fmt.Sprintf("data:%s;base64, %s", contentType, encoded)
			}

			node.Attr[i] = attribute
		}
	}

	return nil
}

func (s *SingleFile) loadResource(ctx context.Context, val, baseURL string) ([]byte, string, error) {
	if !strings.HasPrefix(val, "http://") && !strings.HasPrefix(val, "https://") {
		var err error
		val, err = url.JoinPath(baseURL, val)
		if err != nil {
			return nil, "", fmt.Errorf("join base path %s and url %s: %w", baseURL, val, err)
		}
		val, err = url.PathUnescape(val)
		if err != nil {
			return nil, "", fmt.Errorf("unescape path %s: %w", val, err)
		}
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, val, nil)
	if err != nil {
		return nil, "", fmt.Errorf("new request: %w", err)
	}

	response, err := s.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("do request: %w", err)
	}

	defer func() {
		if response.Body != nil {
			_ = response.Body.Close()
		}
	}()

	if response.StatusCode != http.StatusOK {
		return []byte{}, "", nil
	}

	raw, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, "", fmt.Errorf("read body: %w", err)
	}

	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(raw)))
	base64.StdEncoding.Encode(encoded, raw)

	return encoded, response.Header.Get("Content-Type"), nil
}

func (s *SingleFile) setCharset(node *html.Node, encoding string) {
	var charsetExists bool

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if child.Data == "meta" {
			for _, attribute := range child.Attr {
				if attribute.Key == "charset" {
					charsetExists = true
				}
			}
		}
	}

	if !charsetExists {
		node.AppendChild(&html.Node{
			Type:     html.ElementNode,
			DataAtom: atom.Meta,
			Data:     "meta",
			Attr: []html.Attribute{
				{
					Key: "charset",
					Val: encoding,
				},
			},
		})
	}
}

func baseURL(val string) string {
	parsed, err := url.Parse(val)
	if err != nil {
		return val
	}

	return fmt.Sprintf("%s://%s", parsed.Scheme, parsed.Host)
}

func getEncoding(response *http.Response) string {
	_, encoding, found := strings.Cut(response.Header.Get("Content-Type"), "charset=")
	if !found {
		return defaultEncoding
	}

	encoding = strings.TrimSpace(encoding)

	return encoding
}