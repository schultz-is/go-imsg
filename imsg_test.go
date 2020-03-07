package imsg

import (
	"bytes"
	"encoding/binary"
	"reflect"
	"testing"
)

type marshalTest struct {
	name              string
	imsg              *IMsg
	littleEndianBytes []byte
	bigEndianBytes    []byte
}

var marshalTests = []marshalTest{
	{
		"Empty imsg",
		&IMsg{},
		[]byte{
			0, 0, 0, 0, // Type
			16, 0, // Length
			0, 0, // Flags
			0, 0, 0, 0, // PeerID
			0, 0, 0, 0, // PID
			// No data
		},
		[]byte{
			0, 0, 0, 0,
			0, 16,
			0, 0,
			0, 0, 0, 0,
			0, 0, 0, 0,
		},
	},
	{
		"Simple imsg",
		&IMsg{
			Type:   0xffeeddcc,
			PeerID: 0xffeeddcc,
			PID:    0xffeeddcc,
			Data:   []byte{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67},
		},
		[]byte{
			0xcc, 0xdd, 0xee, 0xff,
			0x17, 0,
			0, 0,
			0xcc, 0xdd, 0xee, 0xff,
			0xcc, 0xdd, 0xee, 0xff,
			0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
		},
		[]byte{
			0xff, 0xee, 0xdd, 0xcc,
			0, 0x17,
			0, 0,
			0xff, 0xee, 0xdd, 0xcc,
			0xff, 0xee, 0xdd, 0xcc,
			0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
		},
	},
	{
		"Flag persistence",
		&IMsg{
			Type:   0xffeeddcc,
			PeerID: 0xffeeddcc,
			PID:    0xffeeddcc,
			Data:   []byte{0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67},
			flags:  0xffee,
		},
		[]byte{
			0xcc, 0xdd, 0xee, 0xff,
			0x17, 0,
			0xee, 0xff,
			0xcc, 0xdd, 0xee, 0xff,
			0xcc, 0xdd, 0xee, 0xff,
			0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
		},
		[]byte{
			0xff, 0xee, 0xdd, 0xcc,
			0, 0x17,
			0xff, 0xee,
			0xff, 0xee, 0xdd, 0xcc,
			0xff, 0xee, 0xdd, 0xcc,
			0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67,
		},
	},
}

func TestMarshalBinary(t *testing.T) {
	// Store out the determined system endianness before manually manipulating it
	systemEndianness := endianness

	var (
		tt     marshalTest
		result []byte
		err    error
	)

	// First run tests for little endian systems
	endianness = binary.LittleEndian
	for _, tt = range marshalTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err = tt.imsg.MarshalBinary()
			if err != nil {
				t.Error(err)
			}

			if bytes.Compare(result, tt.littleEndianBytes) != 0 {
				t.Fatalf("little endian result (% x) does not match expected output (% x)", result, tt.littleEndianBytes)
			}
		})
	}

	// Next run tests for big endian systems
	endianness = binary.BigEndian
	for _, tt = range marshalTests {
		t.Run(tt.name, func(t *testing.T) {
			result, err = tt.imsg.MarshalBinary()
			if err != nil {
				t.Error(err)
			}

			if bytes.Compare(result, tt.bigEndianBytes) != 0 {
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

	// Restore the determined system endianness
	endianness = systemEndianness
}

func TestReadIMsg(t *testing.T) {
	// Store out the determined system endianness before manually manipulating it
	systemEndianness := endianness

	var (
		tt   marshalTest
		buf  *bytes.Buffer
		imsg *IMsg
		err  error
	)

	// First run tests for little endian systems
	endianness = binary.LittleEndian
	for _, tt = range marshalTests {
		t.Run(tt.name, func(t *testing.T) {
			buf = bytes.NewBuffer(tt.littleEndianBytes)
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
	for _, tt = range marshalTests {
		t.Run(tt.name, func(t *testing.T) {
			buf = bytes.NewBuffer(tt.bigEndianBytes)
			imsg, err = ReadIMsg(buf)
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(imsg, tt.imsg) {
				t.Fatalf("big endian result (% x) does not match expected output (% x)", imsg, tt.imsg)
			}
		})
	}

	// Restore the determined system endianness
	endianness = systemEndianness
}
