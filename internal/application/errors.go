package application

import "fmt"

// ValidationError reports input a use case refused. The transport maps it to a
// client error and names the offending field.
type ValidationError struct {
	Field string
	Err   error
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Err)
}

func (e *ValidationError) Unwrap() error {
	return e.Err
}
