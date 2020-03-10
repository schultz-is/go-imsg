// Copyright (c) 2020 Matt Schultz <schultz@sent.com>. All rights reserved.
// Use of this source code is governed by an ISC license that can be found in
// the LICENSE file.

package imsg

import "fmt"

// ErrDataTooLarge is returned when the provided ancillary data is larger than
// is allowed.
type ErrDataTooLarge struct {
	DataLengthInBytes int
	MaxLengthInBytes  uint16
}

// Error implements the error interface.
func (e *ErrDataTooLarge) Error() string {
	return fmt.Sprintf(
		"imsg: provided data is too large (%d bytes > %d bytes)",
		e.DataLengthInBytes,
		e.MaxLengthInBytes,
	)
}

// ErrLengthOutOfBounds is returned when the length parameter is either smaller
// than the imsg header size or larger than the allowed maximum size.
type ErrLengthOutOfBounds struct {
	LengthInBytes    uint16
	MinLengthInBytes uint16
	MaxLengthInBytes uint16
}

// Error implements the error interface.
func (e *ErrLengthOutOfBounds) Error() string {
	return fmt.Sprintf(
		"imsg: message length (%d bytes) is out of allowed bounds (%d - %d bytes)",
		e.LengthInBytes,
		e.MinLengthInBytes,
		e.MaxLengthInBytes,
	)
}

// ErrInsufficientData is returned when reading an imsg produces less data than
// is expected.
type ErrInsufficientData struct {
	ExpectedBytes uint16
	ReadBytes     int
}

// Error implements the error interface.
func (e *ErrInsufficientData) Error() string {
	return fmt.Sprintf(
		"imsg: insufficient data provided (expected %d bytes, read %d bytes)",
		e.ExpectedBytes,
		e.ReadBytes,
	)
}
