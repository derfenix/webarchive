package processors

import (
	"context"
	"fmt"
	"time"

	"github.com/SebastiaanKlippert/go-wkhtmltopdf"

	"github.com/derfenix/webarchive/entity"
)

func NewPDF() *PDF {
	return &PDF{}
}

type PDF struct{}

func (P *PDF) Process(_ context.Context, url string) ([]entity.File, error) {
	gen, err := wkhtmltopdf.NewPDFGenerator()
	if err != nil {
		return nil, fmt.Errorf("new pdf generator: %w", err)
	}

	gen.Dpi.Set(300)
	gen.PageSize.Set(wkhtmltopdf.PageSizeA4)
	gen.Orientation.Set(wkhtmltopdf.OrientationPortrait)
	gen.Grayscale.Set(false)
	gen.Title.Set(url)

	page := wkhtmltopdf.NewPage(url)
	page.PrintMediaType.Set(true)
	page.JavascriptDelay.Set(200)
	page.LoadMediaErrorHandling.Set("ignore")
	page.FooterRight.Set("[page]")
	page.HeaderLeft.Set(url)
	page.HeaderRight.Set(time.Now().Format(time.DateOnly))
	page.FooterFontSize.Set(10)
	page.Zoom.Set(1)
	page.ViewportSize.Set("1920x1080")

	gen.AddPage(page)

	err = gen.Create()
	if err != nil {
		return nil, fmt.Errorf("create pdf: %w", err)
	}

	file := entity.NewFile("page.pdf", gen.Bytes())

	return []entity.File{file}, nil
}
