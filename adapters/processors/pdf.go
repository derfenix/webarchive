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

func (p *PDF) Process(_ context.Context, url string, cache *entity.Cache) ([]entity.File, error) {
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

	opts := wkhtmltopdf.NewPageOptions()
	opts.PrintMediaType.Set(p.cfg.MediaPrint)
	opts.JavascriptDelay.Set(200)
	opts.DisableJavascript.Set(false)
	opts.LoadErrorHandling.Set("ignore")
	opts.LoadMediaErrorHandling.Set("skip")
	opts.FooterRight.Set("[opts]")
	opts.HeaderLeft.Set(url)
	opts.HeaderRight.Set(time.Now().Format(time.DateOnly))
	opts.FooterFontSize.Set(10)
	opts.Zoom.Set(p.cfg.Zoom)
	opts.ViewportSize.Set(p.cfg.Viewport)
	opts.NoBackground.Set(true)
	opts.DisableLocalFileAccess.Set(false)
	opts.DisableExternalLinks.Set(false)
	opts.DisableInternalLinks.Set(false)

	var page wkhtmltopdf.PageProvider
	if len(cache.Get()) > 0 {
		page = &wkhtmltopdf.PageReader{Input: cache.Reader(), PageOptions: opts}
	} else {
		page = &wkhtmltopdf.Page{Input: url, PageOptions: opts}
	}

	gen.AddPage(page)

	err = gen.Create()
	if err != nil {
		return nil, fmt.Errorf("create pdf: %w", err)
	}

	file := entity.NewFile(p.cfg.Filename, gen.Bytes())

	return []entity.File{file}, nil
}
