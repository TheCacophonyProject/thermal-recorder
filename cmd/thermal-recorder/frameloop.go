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
		size:         size,
		currentIndex: 0,
		frames:       frames}
}

// FileWriter wraps a Writer and provides a convenient way of writing
// a CPTV stream to a disk file.
type FrameLoop struct {
	size         int
	currentIndex int
	frames       []*lepton3.Frame
}

func (fl *FrameLoop) nextFrameFrom(index int) int {
	return (index + 1) % fl.size
}

func (fl *FrameLoop) MoveToNextFrame() *lepton3.Frame {
	fl.currentIndex = fl.nextFrameFrom(fl.currentIndex)
	return fl.CurrentFrame()
}

func (fl *FrameLoop) CurrentFrame() *lepton3.Frame {
	return fl.frames[fl.currentIndex]
}

func (fl *FrameLoop) WriteToFile(writer LeptonFrameWriter) error {

	// start with the oldest frame
	firstIndex := fl.nextFrameFrom(fl.currentIndex)

	// Start with an empty previous frame for a new recording.
	firstFrame := new(lepton3.Frame)

	frame := fl.frames[firstIndex]
	prevFrame := frame

	// write first index
	if err := writer.WriteFrame(firstFrame, frame); err != nil {
		return err
	}

	writeIndex := fl.nextFrameFrom(firstIndex)

	// it never writes the current frame as this will be written as part of the program!!
	for writeIndex != fl.currentIndex {
		prevFrame, frame = frame, fl.frames[writeIndex]
		if err := writer.WriteFrame(prevFrame, frame); err != nil {
			return err
		}
		writeIndex = fl.nextFrameFrom(writeIndex)
	}

	return nil
}
