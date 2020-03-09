package imsg

import "fmt"

// ErrDataTooLarge is returned when the provided ancillary data is larger than
// is allowed.
type ErrDataTooLarge struct {
	DataLength    int
	MaximumLength uint16
}

// Error implements the error interface
func (e *ErrDataTooLarge) Error() string {
	return fmt.Sprintf(
		"imsg: provided data is too large (%d bytes > %d bytes)",
		e.DataLength,
		e.MaximumLength,
	)
}
