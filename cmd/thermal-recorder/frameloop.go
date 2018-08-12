// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"github.com/TheCacophonyProject/lepton3"
)

type LeptonFrameWriter interface {
	WriteFrame(prevFrame, frame *lepton3.Frame) error
}

func NewFrameLoop(size int) *FrameLoop {
	frames := make([]*lepton3.Frame, size)
	for i := range frames {
		frames[i] = new(lepton3.Frame)
	}

	return &FrameLoop{
		size:          size,
		currentIndex:  0,
		frames:        frames,
		orderedFrames: make([]*lepton3.Frame, size),
	}
}

type FrameLoop struct {
	size          int
	currentIndex  int
	frames        []*lepton3.Frame
	orderedFrames []*lepton3.Frame
	zeroFrame     lepton3.Frame
}

func (fl *FrameLoop) nextIndexAfter(index int) int {
	return (index + 1) % fl.size
}

func (fl *FrameLoop) Move() *lepton3.Frame {
	fl.currentIndex = fl.nextIndexAfter(fl.currentIndex)
	return fl.Current()
}

func (fl *FrameLoop) Current() *lepton3.Frame {
	return fl.frames[fl.currentIndex]
}

func (fl *FrameLoop) Previous() *lepton3.Frame {
	previousIndex := (fl.currentIndex - 1 + fl.size) % fl.size
	return fl.frames[previousIndex]
}

func (fl *FrameLoop) GetHistory() []*lepton3.Frame {
	// start with the oldest frame
	writeIndex := 0
	readIndex := fl.nextIndexAfter(fl.currentIndex)

	for {
		frame := fl.frames[readIndex]
		if (readIndex < fl.currentIndex) || (*frame != fl.zeroFrame) {
			fl.orderedFrames[writeIndex] = frame
			writeIndex++
		}
		if readIndex == fl.currentIndex {
			return fl.orderedFrames[:writeIndex]
		}

		readIndex = fl.nextIndexAfter(readIndex)
	}

}
