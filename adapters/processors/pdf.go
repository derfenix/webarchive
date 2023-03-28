package processors

import (
	"context"
	"fmt"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"

	"github.com/derfenix/webarchive/config"
	"github.com/derfenix/webarchive/entity"
)

func NewPDF(cfg config.PDF) *PDF {
	return &PDF{cfg: cfg}
}

type PDF struct {
	cfg config.PDF
}

func (p *PDF) Process(_ context.Context, url string) ([]entity.File, error) {
	gen, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("new pdf generator: %w", err)
	}

	gen.Dpi.Set(p.cfg.DPI)
	gen.PageSize.Set(wkhtmltopdf.PageSizeA4)

	if p.cfg.Landscape {
		gen.Orientation.Set(wkhtmltopdf.OrientationLandscape)
	} else {
		gen.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	}

	gen.Grayscale.Set(p.cfg.Grayscale)
	gen.Title.Set(url)

	page := wkhtmltopdf.NewPage(url)
	page.PrintMediaType.Set(p.cfg.MediaPrint)
	page.JavascriptDelay.Set(200)
	page.LoadMediaErrorHandling.Set("ignore")
	page.FooterRight.Set("[page]")
	page.HeaderLeft.Set(url)
	page.HeaderRight.Set(time.Now().Format(time.DateOnly))
	page.FooterFontSize.Set(10)
	page.Zoom.Set(p.cfg.Zoom)
	page.ViewportSize.Set(p.cfg.Viewport)

	gen.AddPage(page)

	err = gen.Create()
	if err != nil {
		return nil, fmt.Errorf("create pdf: %w", err)
	}

	file := entity.NewFile(p.cfg.Filename, gen.Bytes())

	return []entity.File{file}, nil
}
