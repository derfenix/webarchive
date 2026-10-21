package processors

import (
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"

	"github.com/derfenix/webarchive/internal/domain"
)

// newFile detects the MIME type of the archived content. Detection lives here,
// in the adapter, so the domain stays free of content sniffing.
func newFile(name string, data []byte) domain.File {
	return domain.NewFile(uuid.New(), name, mimetype.Detect(data).String(), data, time.Now())
}
