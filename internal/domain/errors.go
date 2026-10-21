package domain

import "errors"

// ErrNotFound reports that a page or a file of a page does not exist.
var ErrNotFound = errors.New("not found")
