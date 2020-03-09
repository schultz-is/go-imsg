// Copyright (c) 2020 Matt Schultz <schultz@sent.com>. All rights reserved.
// Use of this source code is governed by an ISC license that can be found in
// the LICENSE file.

package imsg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"reflect"
	"testing"
)

type imsgTest struct {
	name              string
	imsg              *IMsg
	littleEndianBytes []byte
	bigEndianBytes    []byte
}

var imsgTests = []imsgTest{
	{"empty imsg", &IMsg{}, []byte{0, 0, 0, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, []byte{0, 0, 0, 0, 0, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}},
	{"simple imsg", &IMsg{0xff, 0xee, 0xdd, []byte("test"), 0xcc}, []byte{0xff, 0, 0, 0, 20, 0, 0xcc, 0, 0xee, 0, 0, 0, 0xdd, 0, 0, 0, 0x74, 0x65, 0x73, 0x74}, []byte{0, 0, 0, 0xff, 0, 20, 0, 0xcc, 0, 0, 0, 0xee, 0, 0, 0, 0xdd, 0x74, 0x65, 0x73, 0x74}},
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

func TestMarshalBinary(t *testing.T) {
	// Store out the determined system endianness before manually manipulating it
	systemEndianness := endianness

	var (
		tt     imsgTest
		result []byte
		err    error
		edtl   *ErrDataTooLarge
	)

	// First run tests for little endian systems
	endianness = binary.LittleEndian
	for _, tt = range imsgTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err = tt.imsg.MarshalBinary()
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result, tt.littleEndianBytes) {
				t.Fatalf("little endian result (% x) does not match expected output (% x)", result, tt.littleEndianBytes)
			}
		})
	}

	// Next run tests for big endian systems
	endianness = binary.BigEndian
	for _, tt = range imsgTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err = tt.imsg.MarshalBinary()
			if err != nil {
				t.Error(err)
			}

			if !bytes.Equal(result, tt.bigEndianBytes) {
				t.Fatalf("big endian result (% x) does not match expected output (% x)", result, tt.bigEndianBytes)
			}
		})
	}

	// Ensure imsgs that are too large can't be marshalled
	imsg := &IMsg{
		Data: make([]byte, MaxSizeInBytes),
	}
	_, err = imsg.MarshalBinary()
	if err == nil {
		t.Fatalf("incorrectly marshalled an imsg with oversized ancillary data")
	}
	if !errors.As(err, &edtl) {
		t.Fatalf("failed to marshal an imsg in an unexpected way: %s", err)
	}

	// Restore the determined system endianness
	endianness = systemEndianness
}

func TestReadIMsg(t *testing.T) {
	// Store out the determined system endianness before manually manipulating it
	systemEndianness := endianness

	var (
		tt    imsgTest
		buf   *bytes.Reader
		imsg  *IMsg
		err   error
		eloob *ErrLengthOutOfBounds
		eid   *ErrInsufficientData
	)

	// First run tests for little endian systems
	endianness = binary.LittleEndian
	for _, tt = range imsgTests {
		t.Run(tt.name, func(t *testing.T) {
			buf = bytes.NewReader(tt.littleEndianBytes)
			imsg, err = ReadIMsg(buf)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(imsg, tt.imsg) {
				t.Fatalf("little endian result (% x) does not match expected output (% x)", imsg, tt.imsg)
			}
		})
	}

	// Next run tests for big endian systems
	endianness = binary.BigEndian
	for _, tt = range imsgTests {
		t.Run(tt.name, func(t *testing.T) {
			buf = bytes.NewReader(tt.bigEndianBytes)
			imsg, err = ReadIMsg(buf)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(imsg, tt.imsg) {
				t.Fatalf("big endian result (% x) does not match expected output (% x)", imsg, tt.imsg)
			}
		})
	}

	// Ensure imsgs that have an invalid length aren't unmarshalled
	buf = bytes.NewReader(
		[]byte{
			0, 0, 0, 0, // type
			0, 0, // length < header size
			0, 0, // flags
			0, 0, 0, 0, // peer id
			0, 0, 0, 0, // pid
		},
	)
	_, err = ReadIMsg(buf)
	if err == nil {
		t.Fatalf("incorrectly read an imsg with invalidly small length")
	}
	if !errors.As(err, &eloob) {
		t.Fatalf("failed to read an imsg in an unexpected way: %s", err)
	}

	buf = bytes.NewReader(
		[]byte{
			0, 0, 0, 0, // type
			0xff, 0xff, // length > maximum size
			0, 0, // flags
			0, 0, 0, 0, // peer id
			0, 0, 0, 0, // pid
		},
	)
	_, err = ReadIMsg(buf)
	if err == nil {
		t.Fatalf("incorrectly read an imsg with invalidly large length")
	}
	if !errors.As(err, &eloob) {
		t.Fatalf("failed to read an imsg in an unexpected way: %s", err)
	}

	buf = bytes.NewReader(
		[]byte{
			0, 0, 0, 0,
			0, 0xff,
			0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,

			1, 2, 3, 4,
		},
	)
	_, err = ReadIMsg(buf)
	if err == nil {
		t.Fatalf("incorrectly read an imsg with invalidly short ancillary data")
	}
	if !errors.As(err, &eid) {
		t.Fatalf("failed to read an imsg in an unexpected way: %s", err)
	}

	// Ensure messages smaller than the header size don't get unmershalled
	buf = bytes.NewReader(
		[]byte{0, 0, 0}, // smaller than the uint32 that describes the Type field
	)
	_, err = ReadIMsg(buf)
	if err == nil {
		t.Fatalf("incorrectly read a malformed imsg")
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
