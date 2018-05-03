// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package lepton3

import (
	"fmt"
)

func newFrameBuilder() *frameBuilder {
	f := &frameBuilder{
		segmentPackets: make([][]byte, packetsPerSegment),
		framePackets:   make([][]byte, packetsPerFrame),
	}
	f.reset()
	return f
}

type frameBuilder struct {
	packetNum      int
	segmentNum     int
	segmentPackets [][]byte
	framePackets   [][]byte
}

func (f *frameBuilder) reset() {
	f.packetNum = -1
	f.segmentNum = 0
}

func (f *frameBuilder) nextPacket(packetNum int, packet []byte) (bool, error) {
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

func (f *frameBuilder) sequential(packetNum int) bool {
	if packetNum == 0 && f.packetNum == maxPacketNum {
		return true
	}
	return packetNum == f.packetNum+1
}

func (f *frameBuilder) output(outFrame *RawFrame) {
	for packetNum, packet := range f.framePackets {
		copy(outFrame[packetNum*vospiDataSize:], packet)
	}
}
