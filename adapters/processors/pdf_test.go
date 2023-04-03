//go:build integration

package processors

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPDF_Process(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skip test with external resource")
	}

	files, err := (&PDF{}).Process(context.Background(), "https://github.com/SebastiaanKlippert/go-wkhtmltopdf")
	require.NoError(t, err)
	require.Len(t, files, 1)

	f := files[0]
	fmt.Println("ID         ", f.ID)
	fmt.Println("Name       ", f.Name)
	fmt.Println("MimeType   ", f.MimeType)
	fmt.Println("Size       ", f.Size)
	fmt.Println("Created    ", f.Created.Format(time.RFC3339))
}
