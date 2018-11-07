// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package lepton3

import (
	"fmt"
)

func newFrameBuilder() *frameBuilder {
	f := &frameBuilder{
		segmentBuf: make([]byte, packetsPerSegment*vospiDataSize),
		frameBuf:   make([]byte, packetsPerFrame*vospiDataSize),
	}
	f.reset()
	return f
}

type frameBuilder struct {
	segmentBuf []byte
	frameBuf   []byte
	packetNum  int
	segmentNum int
}

func (f *frameBuilder) reset() {
	f.frameBuf = f.frameBuf[:0]
	f.packetNum = -1
	f.segmentNum = 0
}

func (f *frameBuilder) nextPacket(packetNum int, packet []byte) (bool, error) {
	if !f.sequential(packetNum) {
		return false, fmt.Errorf("out of order packet: %d -> %d", f.packetNum, packetNum)
	}

	copy(f.segmentBuf[packetNum*vospiDataSize:], packet[vospiHeaderSize:])

	switch packetNum {
	case segmentPacketNum:
		// This is the packet that has the segment number set.
		segmentNum := int(packet[0] >> 4)
		if segmentNum > 4 {
			return false, fmt.Errorf("invalid segment number: %d", segmentNum)
		}
		if segmentNum > 0 && segmentNum != f.segmentNum+1 {
			// TODO this might not warrant a resync but certainly ignoring of the segment
			return false, fmt.Errorf("out of order segment")
		}
		f.segmentNum = segmentNum
	case maxPacketNum:
		// End of segment.
		if f.segmentNum > 0 {
			f.frameBuf = append(f.frameBuf, f.segmentBuf...)
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
	copy(outFrame[:], f.frameBuf)
}
