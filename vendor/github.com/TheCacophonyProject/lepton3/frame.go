// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package lepton3

import (
	"encoding/binary"
)

// RawFrame hold the raw bytes for single frame. This is helpful for
// transferring frames around. It can be converted to the more useful
// Frame.
type RawFrame [2 * FrameRows * FrameCols]byte

// Frame represents the thermal readings for a single frame.
type Frame [FrameRows][FrameCols]uint16

// ToFrame converts a RawFrame to a Frame.
func (rf *RawFrame) ToFrame(out *Frame) {
	i := 0
	for y := 0; y < FrameRows; y++ {
		for x := 0; x < FrameCols; x++ {
			out[y][x] = binary.BigEndian.Uint16(rf[i : i+2])
			i += 2
		}
	}
}

// Copy sets current frame as other frame
func (fr *Frame) Copy(orig *Frame) {
	for y := 0; y < FrameRows; y++ {
		copy(fr[y][:], orig[y][:])
	}
}
