// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

const SMALL_SIZE = 3
const LARGE_SIZE = 6

func createAndPopulateFrameLoop(size int) *FrameLoop {
	frameLoop := NewFrameLoop(size)

	frameLoop.Current()[0][0] = 1
	frameLoop.Move()[0][0] = 2
	frameLoop.Move()[0][0] = 3
	return frameLoop
}

func getFrameIds(frames []*lepton3.Frame) []int {
	ids := make([]int, len(frames))
	for ii, frame := range frames {
		ids[ii] = getId(frame)
	}
	return ids
}

func getId(frame *lepton3.Frame) int {
	return int(frame[0][0])
}

func TestFrameLoopPrevious(t *testing.T) {
	frameLoop := createAndPopulateFrameLoop(SMALL_SIZE)

	assert.Equal(t, 2, getId(frameLoop.Previous()))

	frameLoop.Move()[0][0] = 4
	assert.Equal(t, 3, getId(frameLoop.Previous()))
}

func TestFrameLoopLoopsRoundFrames(t *testing.T) {
	frameLoop := createAndPopulateFrameLoop(SMALL_SIZE)
	assert.Equal(t, 1, getId(frameLoop.Move()))
	assert.Equal(t, 2, getId(frameLoop.Move()))
	assert.Equal(t, 3, getId(frameLoop.Move()))
	assert.Equal(t, 1, getId(frameLoop.Move()))
	assert.Equal(t, 2, getId(frameLoop.Move()))
}

func TestFrameLoopHistoryFromStart(t *testing.T) {
	// test write from start of loop
	frameLoop := createAndPopulateFrameLoop(SMALL_SIZE)

	assert.Equal(t, []int{1, 2, 3}, getFrameIds(frameLoop.GetHistory()))
}

func TestFrameLoopHistoryFromMiddleOfLoop(t *testing.T) {
	// test write from start of loop
	frameLoop := createAndPopulateFrameLoop(SMALL_SIZE)
	frameLoop.Move()[0][0] = 4
	frameLoop.Move()[0][0] = 5

	assert.Equal(t, []int{3, 4, 5}, getFrameIds(frameLoop.GetHistory()))
}

func TestFrameLoopHistoryWhenLoopNotFullDoesNotIncludeZeroFrames(t *testing.T) {
	// test write from start of loop
	frameLoop := createAndPopulateFrameLoop(LARGE_SIZE)
	frameLoop.Move()[0][0] = 4

	assert.Equal(t, []int{1, 2, 3, 4}, getFrameIds(frameLoop.GetHistory()))
}
