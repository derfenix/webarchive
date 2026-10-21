// Package domain holds the archive model: pages, their formats, results and
// files. It depends on the standard library only, so the rules below can be
// read and tested without a database, an HTTP client or a logger.
package domain

type Format uint8

const (
	FormatHeaders Format = iota
	FormatSingleFile
	FormatPDF
)

// AllFormats returns a fresh slice: callers may sort or mutate it without
// affecting other callers.
func AllFormats() []Format {
	return []Format{
		FormatHeaders,
		FormatPDF,
		FormatSingleFile,
	}
}
