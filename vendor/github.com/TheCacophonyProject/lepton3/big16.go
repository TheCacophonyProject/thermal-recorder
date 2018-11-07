// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

// Source code in this file comes from the periph project
// (https://periph.io/).

package lepton3

import "encoding/binary"

// Big16 translates big endian 16bits words but everything larger is in little
// endian.
//
// It implements binary.ByteOrder.
var Big16 big16

type big16 struct{}

func (big16) Uint16(b []byte) uint16 {
	_ = b[1] // bounds check hint to compiler; see golang.org/issue/14808
	return uint16(b[1]) | uint16(b[0])<<8
}

func (big16) PutUint16(b []byte, v uint16) {
	_ = b[1] // early bounds check to guarantee safety of writes below
	b[0] = byte(v >> 8)
	b[1] = byte(v)
}

func (big16) Uint32(b []byte) uint32 {
	_ = b[3] // bounds check hint to compiler; see golang.org/issue/14808
	return uint32(b[1]) | uint32(b[0])<<8 | uint32(b[3])<<16 | uint32(b[2])<<24
}

func (big16) PutUint32(b []byte, v uint32) {
	_ = b[3] // early bounds check to guarantee safety of writes below
	b[1] = byte(v)
	b[0] = byte(v >> 8)
	b[3] = byte(v >> 16)
	b[2] = byte(v >> 24)
}

func (big16) Uint64(b []byte) uint64 {
	_ = b[7] // bounds check hint to compiler; see golang.org/issue/14808
	return uint64(b[1]) | uint64(b[0])<<8 | uint64(b[3])<<16 | uint64(b[2])<<24 |
		uint64(b[5])<<32 | uint64(b[4])<<40 | uint64(b[7])<<48 | uint64(b[6])<<56
}

func (big16) PutUint64(b []byte, v uint64) {
	_ = b[7] // early bounds check to guarantee safety of writes below
	b[1] = byte(v)
	b[0] = byte(v >> 8)
	b[3] = byte(v >> 16)
	b[2] = byte(v >> 24)
	b[5] = byte(v >> 32)
	b[4] = byte(v >> 40)
	b[7] = byte(v >> 48)
	b[6] = byte(v >> 56)
}

func (big16) String() string {
	return "big16"
}

var _ binary.ByteOrder = Big16
