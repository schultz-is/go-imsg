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

// This is a fixed-size header used to simplify marshaling and unmarshaling.
type imsgHeader struct {
	Type   uint32
	Length uint16
	Flags  uint16
	PeerID uint32
	PID    uint32
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
// ancillary data is too large, an error is returned. When composing an IMsg
// using this function, the PID field is filled in automatically by a call to
// os.Getpid(). This can be overwritten as desired.
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

	var hdr imsgHeader
	err := binary.Read(r, endianness, &hdr)
	if err != nil {
		return nil, err
	}

	im.Type = hdr.Type
	im.PeerID = hdr.PeerID
	im.PID = hdr.PID
	im.flags = hdr.Flags

	if hdr.Length > HeaderSizeInBytes {
		im.Data = make([]byte, hdr.Length-HeaderSizeInBytes)
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

	if im.Len() > MaxSizeInBytes {
		return nil, errors.New("imsg: imsg exceeds maximum length")
	}

	hdr := imsgHeader{
		Type:   im.Type,
		Length: uint16(im.Len()),
		Flags:  im.flags,
		PeerID: im.PeerID,
		PID:    im.PID,
	}

	err := binary.Write(&buf, endianness, hdr)
	if err != nil {
		return nil, err
	}

	_, err = buf.Write(im.Data)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// SystemEndianness returns the determined system byte order.
func SystemEndianness() binary.ByteOrder {
	return endianness
}
