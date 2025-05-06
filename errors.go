package hostinfo

import "fmt"

// An InvalidLineError is returned by parsers encountering invalid input.
type InvalidLineError struct {
	Prefix, Line string
}

// Error implements the error interface.
func (e *InvalidLineError) Error() string {
	return fmt.Sprintf("%s: %q: invalid line", e.Prefix, e.Line)
}
