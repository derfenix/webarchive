package entity

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Processor interface {
	Process(ctx context.Context, format Format, url string, cache *Cache) Result
	GetMeta(ctx context.Context, url string, cache *Cache) (Meta, error)
}

type Format uint8

const (
	FormatHeaders Format = iota
	FormatSingleFile
	FormatPDF
)

var AllFormats = []Format{
	FormatHeaders,
	FormatPDF,
	FormatSingleFile,
}

type Status uint8

const (
	StatusNew Status = iota
	StatusProcessing
	StatusDone
	StatusFailed
	StatusWithErrors
)

type Meta struct {
	Title       string
	Description string
	Encoding    string
	Error       string
}

type PageBase struct {
	ID          uuid.UUID
	URL         string
	Description string
	Created     time.Time
	Formats     []Format
	Version     uint16
	Status      Status
	Meta        Meta
}

func NewPage(url string, description string, formats ...Format) *Page {
	return &Page{
		PageBase: PageBase{
			ID:          uuid.New(),
			URL:         url,
			Description: description,
			Formats:     formats,
			Created:     time.Now(),
			Version:     1,
		},
		cache: NewCache(),
	}
}

type Page struct {
	PageBase
	Results ResultsRO
	cache   *Cache
}

func (p *Page) SetProcessing() {
	p.Status = StatusProcessing
}

func (p *Page) Process(ctx context.Context, processor Processor) {
	innerWG := sync.WaitGroup{}
	innerWG.Add(len(p.Formats))

	meta, err := processor.GetMeta(ctx, p.URL, p.cache)
	if err != nil {
		p.Meta.Error = err.Error()
	} else {
		p.Meta = meta
	}

	results := Results{}

	for _, format := range p.Formats {
		go func(format Format) {
			defer innerWG.Done()

			defer func() {
				if err := recover(); err != nil {
					results.Add(Result{Format: format, Err: fmt.Errorf("recovered from panic: %v", err)})
				}
			}()

			result := processor.Process(ctx, format, p.URL, p.cache)
			results.Add(result)
		}(format)
	}

	innerWG.Wait()

	var hasResultWithOutErrors bool
	for _, result := range results.Results() {
		if result.Err != nil {
			p.Status = StatusWithErrors
		} else {
			hasResultWithOutErrors = true
		}
	}

	if !hasResultWithOutErrors {
		p.Status = StatusFailed
	}

	if p.Status == StatusProcessing {
		p.Status = StatusDone
	}

	p.Results = results.RO()
}
