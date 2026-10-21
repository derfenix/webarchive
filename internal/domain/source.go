package domain

import (
	"bytes"
	"io"
)

// Source is the page content fetched once and reused by several formats. It is
// a value: processors may read it concurrently and never mutate it.
type Source struct {
	URL     string
	Content []byte
}

func (s Source) HasContent() bool {
	return len(s.Content) > 0
}

func (s Source) Reader() io.Reader {
	return bytes.NewReader(s.Content)
}
