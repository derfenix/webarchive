package domain

// ResultsRO is the stored, read-only list of results of a page. It is a plain
// slice so the storage adapter can encode it as it is today.
type ResultsRO []Result

type Result struct {
	Format Format
	Err    error
	Files  []File
}
