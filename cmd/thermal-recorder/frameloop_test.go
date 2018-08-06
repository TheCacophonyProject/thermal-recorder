// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

const SIZE = 3

var writesRecord [SIZE][2]int
var writes = 0

type TestWriter struct {
}

func (tw TestWriter) WriteFrame(prevFrame, frame *lepton3.Frame) error {
	writesRecord[writes][0] = int(prevFrame[0][0])
	writesRecord[writes][1] = int(frame[0][0])
	writes++
	return nil
}

func createAndPopulateFrameLoop() *FrameLoop {
	frameLoop := NewFrameLoop(SIZE)

	frameLoop.CurrentFrame()[0][0] = 1
	frameLoop.MoveToNextFrame()[0][0] = 2
	frameLoop.MoveToNextFrame()[0][0] = 3
	return frameLoop
}

func TestFrameLoopLoopsRoundFrames(t *testing.T) {
	frameLoop := createAndPopulateFrameLoop()
	assert.Equal(t, uint16(1), frameLoop.MoveToNextFrame()[0][0])
	assert.Equal(t, uint16(2), frameLoop.MoveToNextFrame()[0][0])
	assert.Equal(t, uint16(3), frameLoop.MoveToNextFrame()[0][0])
	assert.Equal(t, uint16(1), frameLoop.MoveToNextFrame()[0][0])
	assert.Equal(t, uint16(2), frameLoop.MoveToNextFrame()[0][0])
}

func TestFrameLoopWriteFromStart(t *testing.T) {
	// test write from start of loop
	frameLoop := createAndPopulateFrameLoop()

	frameLoop.WriteToFile(TestWriter{})
	assert.Equal(t, [2]int{0, 1}, writesRecord[0])
	assert.Equal(t, [2]int{1, 2}, writesRecord[1])

	// now test from middle of loop

	writes = 0
	frameLoop.MoveToNextFrame()[0][0] = 4
	frameLoop.MoveToNextFrame()[0][0] = 5

	frameLoop.WriteToFile(TestWriter{})
	assert.Equal(t, 2, writes)
	assert.Equal(t, [2]int{0, 3}, writesRecord[0])
	assert.Equal(t, [2]int{3, 4}, writesRecord[1])

	// now test from end of loop
	writes = 0
	frameLoop.MoveToNextFrame()[0][0] = 6
	frameLoop.WriteToFile(TestWriter{})

	assert.Equal(t, [2]int{0, 4}, writesRecord[0])
	assert.Equal(t, [2]int{4, 5}, writesRecord[1])
}
