package imsg

import "fmt"

// ErrDataTooLarge is returned when the provided ancillary data is larger than
// is allowed.
type ErrDataTooLarge struct {
	DataLength int
	MaxLength  uint16
}

// Error implements the error interface.
func (e *ErrDataTooLarge) Error() string {
	return fmt.Sprintf(
		"imsg: provided data is too large (%d bytes > %d bytes)",
		e.DataLength,
		e.MaxLength,
	)
}

// ErrLengthOutOfBounds is returned when the length parameter is either smaller
// than the imsg header size or larger than the allowed maximum size.
type ErrLengthOutOfBounds struct {
	Length    uint16
	MinLength uint16
	MaxLength uint16
}

// Error implements the error interface.
func (e *ErrLengthOutOfBounds) Error() string {
	return fmt.Sprintf(
		"imsg: message length (%d bytes) is out of allowed bounds (%d - %d bytes)",
		e.Length,
		e.MinLength,
		e.MaxLength,
	)
}
