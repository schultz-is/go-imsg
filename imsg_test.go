// Copyright (c) 2020 Matt Schultz <schultz@sent.com>. All rights reserved.
// Use of this source code is governed by an ISC license that can be found in
// the LICENSE file.

package imsg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
)

type imsgTest struct {
	name              string
	imsg              *IMsg
	littleEndianBytes []byte
	bigEndianBytes    []byte
	expectedErrorType error
}

var marshalTests = []imsgTest{
	{"valid empty", &IMsg{}, []byte{0, 0, 0, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, []byte{0, 0, 0, 0, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil},
	{"valid simple", &IMsg{0xff, 0xee, 0xdd, []byte("test"), 0xcc}, []byte{0xff, 0, 0, 0, 20, 0, 0xcc, 0, 0xee, 0, 0, 0, 0xdd, 0, 0, 0, 0x74, 0x65, 0x73, 0x74}, []byte{0, 0, 0, 0xff, 0, 20, 0, 0xcc, 0, 0, 0, 0xee, 0, 0, 0, 0xdd, 0x74, 0x65, 0x73, 0x74}, nil},
	{"invalid data too large", &IMsg{Data: make([]byte, MaxSizeInBytes+1)}, nil, nil, &ErrDataTooLarge{}},
}

var unmarshalTests = []imsgTest{
	{"valid empty", &IMsg{}, []byte{0, 0, 0, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, []byte{0, 0, 0, 0, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, nil},
	{"valid simple", &IMsg{0xff, 0xee, 0xdd, []byte("test"), 0xcc}, []byte{0xff, 0, 0, 0, 20, 0, 0xcc, 0, 0xee, 0, 0, 0, 0xdd, 0, 0, 0, 0x74, 0x65, 0x73, 0x74}, []byte{0, 0, 0, 0xff, 0, 20, 0, 0xcc, 0, 0, 0, 0xee, 0, 0, 0, 0xdd, 0x74, 0x65, 0x73, 0x74}, nil},
	{"invalid < min length", nil, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, &ErrLengthOutOfBounds{}},
	{"invalid > max length", nil, []byte{0, 0, 0, 0, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, []byte{0, 0, 0, 0, 0xff, 0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, &ErrLengthOutOfBounds{}},
	{"invalid insufficient data", nil, []byte{0, 0, 0, 0, 0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, []byte{0, 0, 0, 0, 0, 0xff, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, &ErrInsufficientData{}},
	{"invalid insufficient data", nil, []byte{0, 0, 0}, []byte{0, 0, 0}, io.ErrUnexpectedEOF},
}

func TestComposeIMsg(t *testing.T) {
	var edtl *ErrDataTooLarge

	// Assemble a valid imsg
	_, err := ComposeIMsg(0, 0, nil)
	if err != nil {
		t.Fatalf("failed to compose valid imsg")
	}

	// Assemble an invalid imsg
	_, err = ComposeIMsg(0, 0, make([]byte, MaxSizeInBytes+1))
	if err == nil {
		t.Fatalf("incorrectly composed an invalid imsg")
	}
	if !errors.As(err, &edtl) {
		t.Fatalf("failed to compose an imsg in an unexpected way: %s", err)
	}
}

func TestReadIMsg(t *testing.T) {
	// Store out the determined system endianness before manually manipulating it
	systemEndianness := endianness

	var (
		tt     imsgTest
		buf    *bytes.Reader
		result *IMsg
		err    error
	)

	// First run tests for little endian systems
	endianness = binary.LittleEndian
	for _, tt = range unmarshalTests {
		t.Run(
			fmt.Sprintf("%s little endian", tt.name),
			func(t *testing.T) {
				buf = bytes.NewReader(tt.littleEndianBytes)
				result, err = ReadIMsg(buf)

				if tt.expectedErrorType == nil {
					if err != nil {
						t.Fatalf("unexpected ReadIMsg failure: %s", err)
					}

					if !reflect.DeepEqual(result, tt.imsg) {
						t.Fatalf("result of ReadIMsg does not match expected output (%#v != %#v)", result, tt.imsg)
					}
				} else {
					if err == nil {
						t.Fatalf("incorrectly read imsg")
					}

					if !errors.As(err, &tt.expectedErrorType) {
						t.Fatalf("failed to read imsg in unexpected way: %s", err)
					}
				}
			})
	}

	// Next run tests for big endian systems
	endianness = binary.BigEndian
	for _, tt = range unmarshalTests {
		t.Run(
			fmt.Sprintf("%s big endian", tt.name),
			func(t *testing.T) {
				buf = bytes.NewReader(tt.bigEndianBytes)
				result, err = ReadIMsg(buf)

				if tt.expectedErrorType == nil {
					if err != nil {
						t.Fatalf("unexpected ReadIMsg failure: %s", err)
					}

					if !reflect.DeepEqual(result, tt.imsg) {
						t.Fatalf("result of ReadIMsg does not match expected output (%#v != %#v)", result, tt.imsg)
					}
				} else {
					if err == nil {
						t.Fatalf("incorrectly read imsg")
					}

					if !errors.As(err, &tt.expectedErrorType) {
						t.Fatalf("failed to read imsg in unexpected way: %s", err)
					}
				}
			})
	}

	// Restore the determined system endianness
	endianness = systemEndianness
}

func TestUnmarshalBinary(t *testing.T) {
	// Store out the determined system endianness before manually manipulating it
	systemEndianness := endianness

	var (
		tt     imsgTest
		result *IMsg
		err    error
	)

	// First run tests for little endian systems
	endianness = binary.LittleEndian
	for _, tt = range unmarshalTests {
		t.Run(
			fmt.Sprintf("%s little endian", tt.name),
			func(t *testing.T) {
				result = &IMsg{}
				err = result.UnmarshalBinary(tt.littleEndianBytes)

				if tt.expectedErrorType == nil {
					if err != nil {
						t.Fatalf("unexpected ReadIMsg failure: %s", err)
					}

					if !reflect.DeepEqual(result, tt.imsg) {
						t.Fatalf("result of ReadIMsg does not match expected output (%#v != %#v)", result, tt.imsg)
					}
				} else {
					if err == nil {
						t.Fatalf("incorrectly read imsg")
					}

					if !errors.As(err, &tt.expectedErrorType) {
						t.Fatalf("failed to read imsg in unexpected way: %s", err)
					}
				}
			})
	}

	// Next run tests for big endian systems
	endianness = binary.BigEndian
	for _, tt = range unmarshalTests {
		t.Run(
			fmt.Sprintf("%s big endian", tt.name),
			func(t *testing.T) {
				result = &IMsg{}
				err = result.UnmarshalBinary(tt.bigEndianBytes)

				if tt.expectedErrorType == nil {
					if err != nil {
						t.Fatalf("unexpected ReadIMsg failure: %s", err)
					}

					if !reflect.DeepEqual(result, tt.imsg) {
						t.Fatalf("result of ReadIMsg does not match expected output (%#v != %#v)", result, tt.imsg)
					}
				} else {
					if err == nil {
						t.Fatalf("incorrectly read imsg")
					}

					if !errors.As(err, &tt.expectedErrorType) {
						t.Fatalf("failed to read imsg in unexpected way: %s", err)
					}
				}
			})
	}

	// Restore the determined system endianness
	endianness = systemEndianness
}

func TestMarshalBinary(t *testing.T) {
	// Store out the determined system endianness before manually manipulating it
	systemEndianness := endianness

	var (
		tt     imsgTest
		result []byte
		err    error
	)

	// First run tests for little endian systems
	endianness = binary.LittleEndian
	for _, tt = range marshalTests {
		t.Run(
			fmt.Sprintf("%s little endian", tt.name),
			func(t *testing.T) {
				result, err = tt.imsg.MarshalBinary()

				if tt.expectedErrorType == nil {
					if err != nil {
						t.Fatalf("unexpected MarshalBinary failure: %s", err)
					}

					if !bytes.Equal(result, tt.littleEndianBytes) {
						t.Fatalf("result of MarshalBinary does not match expected output (% x != % x)", result, tt.littleEndianBytes)
					}
				} else {
					if err == nil {
						t.Fatal("incorrectly marshalled imsg to binary")
					}

					if !errors.As(err, &tt.expectedErrorType) {
						t.Fatalf("failed to marshal imsg to binary in unexpected way: %s", err)
					}
				}
			})
	}

	// Next run tests for big endian systems
	endianness = binary.BigEndian
	for _, tt = range marshalTests {
		t.Run(
			fmt.Sprintf("%s big endian", tt.name),
			func(t *testing.T) {
				result, err = tt.imsg.MarshalBinary()

				if tt.expectedErrorType == nil {
					if err != nil {
						t.Fatalf("unexpected MarshalBinary failure: %s", err)
					}

					if !bytes.Equal(result, tt.bigEndianBytes) {
						t.Fatalf("result of MarshalBinary does not match expected output (% x != % x)", result, tt.bigEndianBytes)
					}
				} else {
					if err == nil {
						t.Fatal("incorrectly marshalled imsg to binary")
					}

					if !errors.As(err, &tt.expectedErrorType) {
						t.Fatalf("failed to marshal imsg to binary in unexpected way: %s", err)
					}
				}
			})
	}

	// Restore the determined system endianness
	endianness = systemEndianness
}

func TestSystemEndianness(t *testing.T) {
	// Store out the determined system endianness before manually manipulating it
	systemEndianness := endianness

	if endianness != SystemEndianness() {
		t.Fatalf("determined endianness does not match expected value")
	}

	endianness = binary.LittleEndian
	if endianness != SystemEndianness() {
		t.Fatalf("determined endianness does not match expected value")
	}

	endianness = binary.BigEndian
	if endianness != SystemEndianness() {
		t.Fatalf("determined endianness does not match expected value")
	}

	// Restore the determined system endianness
	endianness = systemEndianness
}

func TestLen(t *testing.T) {
	imsg := &IMsg{}

	if imsg.Len() != HeaderSizeInBytes {
		t.Fatalf("empty imsg length (%d) should match header length (%d)", imsg.Len(), HeaderSizeInBytes)
	}
}
