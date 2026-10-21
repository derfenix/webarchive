package domain

import (
	"time"

	"github.com/google/uuid"
)

type File struct {
	ID       uuid.UUID
	Name     string
	MimeType string
	Size     int64
	Data     []byte
	Created  time.Time
}

// NewFile builds a file record. The MIME type is provided by the caller: the
// domain does not sniff content.
func NewFile(id uuid.UUID, name, mimeType string, data []byte, created time.Time) File {
	return File{
		ID:       id,
		Name:     name,
		MimeType: mimeType,
		Size:     int64(len(data)),
		Data:     data,
		Created:  created,
	}
}
