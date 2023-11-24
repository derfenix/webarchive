package internal

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/disintegration/imaging"
	"github.com/gabriel-vasile/mimetype"
	"go.uber.org/zap"
	"golang.org/x/net/html"
)

type MediaInline struct {
	log    *zap.Logger
	getter func(context.Context, string) (*http.Response, error)
}

func NewMediaInline(log *zap.Logger, getter func(context.Context, string) (*http.Response, error)) *MediaInline {
	return &MediaInline{log: log, getter: getter}
}

func (m *MediaInline) Inline(ctx context.Context, reader io.Reader, pageURL string) (*html.Node, error) {
	htmlNode, err := html.Parse(reader)
	if err != nil {
		return nil, fmt.Errorf("parse response body: %w", err)
	}

	baseURL, err := url.Parse(pageURL)
	if err != nil {
		return nil, fmt.Errorf("parse page url: %w", err)
	}

	m.visit(ctx, htmlNode, m.processorFunc, baseURL)

	return htmlNode, nil
}

func (m *MediaInline) processorFunc(ctx context.Context, node *html.Node, baseURL *url.URL) error {
	switch node.Data {
	case "link":
		if err := m.processHref(ctx, node.Attr, baseURL); err != nil {
			return fmt.Errorf("process link %s: %w", node.Attr, err)
		}

	case "script", "img":
		if err := m.processSrc(ctx, node.Attr, baseURL); err != nil {
			return fmt.Errorf("process script %s: %w", node.Attr, err)
		}

	case "a":
		if err := m.processAHref(node.Attr, baseURL); err != nil {
			return fmt.Errorf("process a href %s: %w", node.Attr, err)
		}
	}

	return nil
}

func (m *MediaInline) processAHref(attrs []html.Attribute, baseURL *url.URL) error {
	for idx, attr := range attrs {
		switch attr.Key {
		case "href":
			attrs[idx].Val = normalizeURL(attr.Val, baseURL)
		}
	}

	return nil
}

func (m *MediaInline) processHref(ctx context.Context, attrs []html.Attribute, baseURL *url.URL) error {
	var shouldProcess bool
	var value string
	var valueIdx int

	for idx, attr := range attrs {
		switch attr.Key {
		case "rel":
			switch attr.Val {
			case "stylesheet", "icon", "alternate icon", "shortcut icon", "manifest":
				shouldProcess = true
			}

		case "href":
			value = attr.Val
			valueIdx = idx
		}
	}

	if !shouldProcess {
		return nil
	}

	encodedValue, err := m.loadAndEncode(ctx, baseURL, value)
	if err != nil {
		return err
	}

	attrs[valueIdx].Val = encodedValue

	return nil
}

func (m *MediaInline) processSrc(ctx context.Context, attrs []html.Attribute, baseURL *url.URL) error {
	var shouldProcess bool
	var value string
	var valueIdx int

	for idx, attr := range attrs {
		switch attr.Key {
		case "src":
			value = attr.Val
			valueIdx = idx
			shouldProcess = true
		case "data-src":
			value = attr.Val
		}
	}

	if !shouldProcess {
		return nil
	}

	encodedValue, err := m.loadAndEncode(ctx, baseURL, value)
	if err != nil {
		return err
	}

	attrs[valueIdx].Val = encodedValue

	return nil
}

func (m *MediaInline) loadAndEncode(ctx context.Context, baseURL *url.URL, value string) (string, error) {
	mime := "text/plain"

	if value == "" {
		return "", nil
	}

	normalizedURL := normalizeURL(value, baseURL)
	if normalizedURL == "" {
		return value, nil
	}

	response, err := m.getter(ctx, normalizedURL)
	if err != nil {
		m.log.Sugar().With(zap.Error(err)).Errorf("load %s", normalizedURL)
		return value, nil
	}

	defer func() {
		_ = response.Body.Close()
	}()

	cleanMime := func(s string) string {
		s, _, _ = strings.Cut(s, "+")
		return s
	}

	if ct := response.Header.Get("Content-Type"); ct != "" {
		mime = ct
	}

	encodedVal, err := m.encodeResource(response.Body, &mime)
	if err != nil {
		return value, fmt.Errorf("encode resource: %w", err)
	}

	return fmt.Sprintf("data:%s;base64, %s", cleanMime(mime), encodedVal), nil
}

func (m *MediaInline) visit(ctx context.Context, n *html.Node, proc func(context.Context, *html.Node, *url.URL) error, baseURL *url.URL) {
	if err := proc(ctx, n, baseURL); err != nil {
		m.log.Error("process error", zap.Error(err))
	}

	if n.FirstChild != nil {
		m.visit(ctx, n.FirstChild, proc, baseURL)
	}

	if n.NextSibling != nil {
		m.visit(ctx, n.NextSibling, proc, baseURL)
	}
}

func normalizeURL(resourceURL string, base *url.URL) string {
	if strings.HasPrefix(resourceURL, "//") {
		return "https:" + resourceURL
	}

	if strings.HasPrefix(resourceURL, "about:") {
		return ""
	}

	parsedResourceURL, err := url.Parse(resourceURL)
	if err != nil {
		return resourceURL
	}

	reference := base.ResolveReference(parsedResourceURL)

	return reference.String()
}

func (m *MediaInline) encodeResource(r io.Reader, mime *string) (string, error) {
	all, err := io.ReadAll(r)
	if err != nil {
		return "", fmt.Errorf("read data: %w", err)
	}

	all, err = m.preprocessResource(all, mime)
	if err != nil {
		return "", fmt.Errorf("preprocess resource: %w", err)
	}

	return base64.StdEncoding.EncodeToString(all), nil
}

func (m *MediaInline) preprocessResource(data []byte, mime *string) ([]byte, error) {
	detectedMime := mimetype.Detect(data)

	switch {
	case strings.HasPrefix(detectedMime.String(), "image"):
		decodedImage, err := imaging.Decode(bytes.NewBuffer(data))
		if err != nil {
			m.log.Error("failed to decode image", zap.Error(err))

			return data, nil
		}

		if size := decodedImage.Bounds().Size(); size.X > 1024 || size.Y > 1024 {
			thumbnail := imaging.Thumbnail(decodedImage, 1024, 1024, imaging.Lanczos)
			buf := bytes.NewBuffer(nil)

			if err := imaging.Encode(buf, thumbnail, imaging.JPEG, imaging.JPEGQuality(90)); err != nil {
				m.log.Error("failed to create resized image", zap.Error(err))

				return data, nil
			}

			*mime = "image/jpeg"
			m.log.Info("Resized")

			return buf.Bytes(), nil
		}
	}

	return data, nil
}
