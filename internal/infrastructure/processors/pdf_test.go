//go:build integration

package processors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/derfenix/webarchive/internal/config"
	"github.com/derfenix/webarchive/internal/domain"
)

func TestPDF_Process(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skip test with external resource")
	}

	pdf := NewPDF(config.PDF{
		MediaPrint: true,
		Zoom:       1,
		Viewport:   "1280x720",
		DPI:        150,
		Filename:   "page.pdf",
	})

	files, err := pdf.Process(context.Background(), domain.Source{URL: "https://github.com/SebastiaanKlippert/go-wkhtmltopdf"})
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	fmt.Println("ID         ", f.ID)
	fmt.Println("Name       ", f.Name)
	fmt.Println("MimeType   ", f.MimeType)
	fmt.Println("Size       ", f.Size)
	fmt.Println("Created    ", f.Created.Format(time.RFC3339))
}
