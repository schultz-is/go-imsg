// Copyright (c) 2020 Matt Schultz <schultz@sent.com>. All rights reserved.
// Use of this source code is governed by an ISC license that can be found in
// the LICENSE file.

// Package imsg provides tools for working with OpenBSD's imsg library.
package imsg

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"unsafe"
)

const (
	// An imsg header is 16 bytes of information that precede the data being
	// transmitted or received.
	HeaderSizeInBytes = 16
	// Single imsgs should not be larger than the currently defined maximum of
	// 16384 bytes.
	MaxSizeInBytes = 16384
)

// This is the system's endianness, which is used to convert imsgs to and from
// binary.
var endianness binary.ByteOrder

func init() {
	// N.B. This is a gross little hack to determine the system's endianness by
	// creating an unsafe.Pointer of a 16-bit int and directly inspecting its
	// in-memory layout.
	var x uint16 = 1
	bs := (*[2]byte)(unsafe.Pointer(&x))
	if bs[0] == 1 {
		endianness = binary.LittleEndian
	} else {
		endianness = binary.BigEndian
	}
}

// An IMsg is a message used to aid inter-process communication over sockets,
// often when processes with different privileges are required to cooperate.
type IMsg struct {
	Type   uint32 // Describes the meaning of the message
	PeerID uint32 // Free for use by caller; intended to identify message sender
	PID    uint32 // Free for use by caller; intended to identify message sender
	Data   []byte // Ancillary data included with the imsg

	// Flags are used internally by imsg functions in the C implementation and
	// should not be used by applications. For that reason, they're included but
	// unused in this library.
	flags uint16
}

// ComposeIMsg constructs an IMsg of the provided type. If the included
// ancillary data is too large, an error is returned.
func ComposeIMsg(
	typ, peerID uint32,
	data []byte,
) (*IMsg, error) {
	if len(data) > (MaxSizeInBytes - HeaderSizeInBytes) {
		return nil, errors.New("imsg: provided data is too large")
	}

	return &IMsg{
		Type:   typ,
		PeerID: peerID,
		PID:    uint32(os.Getpid()),
		Data:   data,
	}, nil
}

// ReadIMsg constructs an IMsg by reading from an io.Reader. If the incoming
// data is malformed, this function can block by attempting to read more data
// than is present.
func ReadIMsg(r io.Reader) (*IMsg, error) {
	im := &IMsg{}

	err := binary.Read(r, endianness, &(im.Type))
	if err != nil {
		return nil, err
	}

	var msgLen uint16
	err = binary.Read(r, endianness, &msgLen)
	if err != nil {
		return nil, err
	}
	if msgLen < HeaderSizeInBytes || msgLen > MaxSizeInBytes {
		return nil, errors.New("imsg: invalid imsg length received")
	}

	err = binary.Read(r, endianness, &(im.flags))
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, endianness, &(im.PeerID))
	if err != nil {
		return nil, err
	}

	err = binary.Read(r, endianness, &(im.PID))
	if err != nil {
		return nil, err
	}

	if msgLen > HeaderSizeInBytes {
		im.Data = make([]byte, msgLen-HeaderSizeInBytes)
		_, err = r.Read(im.Data)
		if err != nil {
			return nil, err
		}
	}

	return im, nil
}

// Len returns the size in bytes of the imsg.
func (im *IMsg) Len() int {
	return len(im.Data) + HeaderSizeInBytes
}

// MarshalBinary implements the encoding.BinaryMarshaler interface.
func (im IMsg) MarshalBinary() ([]byte, error) {
	var buf bytes.Buffer

	err := binary.Write(&buf, endianness, im.Type)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, endianness, uint16(im.Len()))
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, endianness, im.flags)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, endianness, im.PeerID)
	if err != nil {
		return nil, err
	}

	err = binary.Write(&buf, endianness, im.PID)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(im.Data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
