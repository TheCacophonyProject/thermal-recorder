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
type RawFrame [packetsPerFrame * vospiDataSize]byte

// Frame represents the thermal readings for a single frame.
type Frame struct {
	Pix    [FrameRows][FrameCols]uint16
	Status Telemetry
}

// FramesHz define the approximate number of frames per second emitted by the Lepton 3 camera.
const FramesHz = 9

// ToFrame converts a RawFrame to a Frame.
func (rf *RawFrame) ToFrame(out *Frame) error {
	if err := ParseTelemetry(rf[:], &out.Status); err != nil {
		return err
	}

	rawPix := rf[telemetryPacketCount*vospiDataSize:]
	i := 0
	for y := 0; y < FrameRows; y++ {
		for x := 0; x < FrameCols; x++ {
			out.Pix[y][x] = binary.BigEndian.Uint16(rawPix[i : i+2])
			i += 2
		}
	}
	return nil
}

// Copy sets current frame as other frame
func (fr *Frame) Copy(orig *Frame) {
	fr.Status = orig.Status
	for y := 0; y < FrameRows; y++ {
		copy(fr.Pix[y][:], orig.Pix[y][:])
	}
}
