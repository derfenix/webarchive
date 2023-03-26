package entity

import (
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
)

func NewFile(name string, data []byte) File {
	detected := mimetype.Detect(data)

	return File{
		ID:       uuid.New(),
		Name:     name,
		MimeType: detected.String(),
		Size:     int64(len(data)),
		Data:     data,
		Created:  time.Now(),
	}
}

type File struct {
	ID       uuid.UUID
	Name     string
	MimeType string
	Size     int64
	Data     []byte
	Created  time.Time
}
