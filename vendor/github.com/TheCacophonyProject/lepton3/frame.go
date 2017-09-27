// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package lepton3

import (
	"encoding/binary"
	"fmt"
)

func newFrame() *frame {
	f := &frame{
		segmentPackets: make([][]byte, packetsPerSegment),
		framePackets:   make([][]byte, packetsPerFrame),
	}
	f.reset()
	return f
}

type frame struct {
	packetNum      int
	segmentNum     int
	segmentPackets [][]byte
	framePackets   [][]byte
}

func (f *frame) reset() {
	f.packetNum = -1
	f.segmentNum = 0
}

func (f *frame) nextPacket(packetNum int, packet []byte) (bool, error) {
	if !f.sequential(packetNum) {
		return false, fmt.Errorf("out of order packet: %d -> %d", f.packetNum, packetNum)
	}

	// Store the packet data in current segment
	f.segmentPackets[packetNum] = packet[vospiHeaderSize:]

	switch packetNum {
	case segmentPacketNum:
		segmentNum := int(packet[0] >> 4)
		if segmentNum > 4 {
			return false, fmt.Errorf("invalid segment number: %d", segmentNum)
		}
		if segmentNum > 0 && segmentNum != f.segmentNum+1 {
			// XXX this might not warrant a resync but certainly ignoring of the segment
			return false, fmt.Errorf("out of order segment")
		}
		f.segmentNum = segmentNum
	case maxPacketNum:
		if f.segmentNum > 0 {
			// This should be fast as only slice headers for the
			// segment are being copied, not the packet data itself.
			copy(f.framePackets[(f.segmentNum-1)*packetsPerSegment:], f.segmentPackets)
		}
		if f.segmentNum == 4 {
			// Complete frame!
			return true, nil
		}
	}
	f.packetNum = packetNum
	return false, nil
}

func (f *frame) sequential(packetNum int) bool {
	if packetNum == 0 && f.packetNum == maxPacketNum {
		return true
	}
	return packetNum == f.packetNum+1
}

func (f *frame) output(outFrame *Frame) {
	for packetNum, packet := range f.framePackets {
		for i := 0; i < vospiDataSize; i += 2 {
			x := i >> 1 // divide 2
			if packetNum%2 == 1 {
				x += colsPerPacket
			}
			y := packetNum >> 1 // divide 2
			outFrame[y][x] = binary.BigEndian.Uint16(packet[i : i+2])
		}
	}
}
