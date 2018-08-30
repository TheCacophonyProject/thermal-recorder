// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"sync"

	"github.com/TheCacophonyProject/lepton3"
)

var mu sync.Mutex

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
		bufferFull:    false,
	}
}

// FrameLoop stores the last n frames in a loop that will be overwritten when full.
// The latest written frame can be anywhere in the list of frames.  Beware: all frames
// returned by FrameLoop will at some point be over-written.
type FrameLoop struct {
	size          int
	currentIndex  int
	frames        []*lepton3.Frame
	orderedFrames []*lepton3.Frame
	zeroFrame     lepton3.Frame
	bufferFull    bool
}

func (fl *FrameLoop) nextIndexAfter(index int) int {
	return (index + 1) % fl.size
}

// Move, moves the current frame one forwards and return the new frame.
// Note: data on all returned frame objects will eventually get overwritten
func (fl *FrameLoop) Move() *lepton3.Frame {
	mu.Lock()
	defer mu.Unlock()
	if fl.currentIndex == fl.size-1 {
		fl.bufferFull = true
	}

	fl.currentIndex = fl.nextIndexAfter(fl.currentIndex)
	return fl.Current()
}

// Current returns the current frame.
// Note: data on all returned frame objects will eventually get overwritten
func (fl *FrameLoop) Current() *lepton3.Frame {
	return fl.frames[fl.currentIndex]
}

// Previous returns the previous frame.
// Note: data on all returned frame objects will eventually get overwritten
func (fl *FrameLoop) Previous() *lepton3.Frame {
	mu.Lock()
	defer mu.Unlock()
	if fl == nil {
		return nil
	}
	previousIndex := (fl.currentIndex - 1 + fl.size) % fl.size
	f := new(lepton3.Frame)
	f.Copy(fl.frames[previousIndex])
	return f
}

// GetHistory returns all the frames recorded in an slice from oldest to newest.
// Note: The returned slice will be rewritten next time GetHistory is called.
func (fl *FrameLoop) GetHistory() []*lepton3.Frame {
	mu.Lock()
	defer mu.Unlock()
	if fl.currentIndex == fl.size-1 {
		copy(fl.orderedFrames[:], fl.frames[:])
		return fl.orderedFrames
	}

	nextIndex := fl.nextIndexAfter(fl.currentIndex)

	if !fl.bufferFull {
		copy(fl.orderedFrames, fl.frames[:nextIndex])
		return fl.orderedFrames[:nextIndex]
	}

	copy(fl.orderedFrames, fl.frames[nextIndex:])
	copy(fl.orderedFrames[fl.size-nextIndex:], fl.frames[:nextIndex])
	return fl.orderedFrames
}
