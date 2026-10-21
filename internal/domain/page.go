package domain

import (
	"time"

	"github.com/google/uuid"
)

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

type Page struct {
	PageBase
	Results ResultsRO
}

func NewPage(id uuid.UUID, url, description string, created time.Time, formats ...Format) *Page {
	return &Page{
		PageBase: PageBase{
			ID:          id,
			URL:         url,
			Description: description,
			Formats:     formats,
			Created:     created,
			Version:     1,
			Status:      StatusNew,
		},
	}
}

func (p *Page) StartProcessing() {
	p.Status = StatusProcessing
}

func (p *Page) SetMeta(meta Meta) {
	p.Meta = meta
}

func (p *Page) SetMetaError(err error) {
	p.Meta.Error = err.Error()
}

// Finish stores the per-format results and derives the page status from them:
// any failed format makes the page "with errors", no successful format at all
// makes it "failed", and a page that finished its work without both is "done".
func (p *Page) Finish(results []Result) {
	p.Results = results

	var hasResultWithoutErrors bool

	for _, result := range results {
		if result.Err != nil {
			p.Status = StatusWithErrors
		} else {
			hasResultWithoutErrors = true
		}
	}

	if !hasResultWithoutErrors {
		p.Status = StatusFailed
	}

	if p.Status == StatusProcessing {
		p.Status = StatusDone
	}
}
