// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

const FIVE_FRAME_LOOP = 5

type FrameLoopTestClass struct {
	*FrameLoop
	count uint16
}

func NewFrameLoopTestClass(frames int) *FrameLoopTestClass {
	frameLoop := &FrameLoopTestClass{
		FrameLoop: NewFrameLoop(FIVE_FRAME_LOOP),
		count:     1,
	}
	if frames > 0 {
		frameLoop.Current()[0][0] = frameLoop.count
	}
	frameLoop.AddFrames(frames - 1)
	return frameLoop
}

func (fl *FrameLoopTestClass) Move() *lepton3.Frame {
	frame := fl.FrameLoop.Move()
	fl.count++
	frame[0][0] = fl.count
	return frame
}

func (fl *FrameLoopTestClass) AddFrames(numberFrames int) {
	for ii := 0; ii < numberFrames; ii++ {
		fl.Move()
	}
}

func getId(frame *lepton3.Frame) int {
	return int(frame[0][0])
}

func getHistoryIds(frameLoop *FrameLoopTestClass) []int {
	frames := frameLoop.GetHistory()
	ids := make([]int, len(frames))
	for ii, frame := range frames {
		ids[ii] = getId(frame)
	}
	return ids
}

func TestFrameLoopLoopsRoundFrames(t *testing.T) {
	frameLoop := NewFrameLoop(FIVE_FRAME_LOOP)
	frameLoop.Current()[0][0] = uint16(1)
	frameLoop.Move()[0][0] = uint16(2)
	frameLoop.Move()[0][0] = uint16(3)
	frameLoop.Move()[0][0] = uint16(4)
	frameLoop.Move()[0][0] = uint16(5)

	assert.Equal(t, 1, getId(frameLoop.Move()))
	assert.Equal(t, 2, getId(frameLoop.Move()))
	assert.Equal(t, 3, getId(frameLoop.Move()))
	assert.Equal(t, 4, getId(frameLoop.Move()))
	assert.Equal(t, 5, getId(frameLoop.Move()))
	assert.Equal(t, 1, getId(frameLoop.Move()))
	assert.Equal(t, 2, getId(frameLoop.Move()))
}

func TestFrameHistoryFromStart(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(1)

	frameLoop.SetAsOldest()
	assert.Equal(t, []int{1}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{1, 2}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{1, 2, 3}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{1, 2, 3, 4}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{1, 2, 3, 4, 5}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{2, 3, 4, 5, 6}, getHistoryIds(frameLoop))
	frameLoop.Move()
	frameLoop.Move()
	frameLoop.Move()
	frameLoop.Move()
	assert.Equal(t, []int{6, 7, 8, 9, 10}, getHistoryIds(frameLoop))
}

func TestFrameLoopHistoryDoesNotIncludeUnwrittenFrames(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(2)
	assert.Equal(t, []int{1, 2}, getHistoryIds(frameLoop))
}

func TestFrameLoopHistoryFromEndFirstTime(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(5)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, getHistoryIds(frameLoop))
}

func TestFrameLoopHistoryFromFirstInLoop(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(6)
	assert.Equal(t, []int{2, 3, 4, 5, 6}, getHistoryIds(frameLoop))
}

func TestFrameLoopHistoryFromMiddleOfLoop(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(8)
	assert.Equal(t, []int{4, 5, 6, 7, 8}, getHistoryIds(frameLoop))
}

func TestFrameLoopHistoryFromLastPositionOfLoop(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(10)
	assert.Equal(t, []int{6, 7, 8, 9, 10}, getHistoryIds(frameLoop))
}

func TestFrameLoopHistoryWhenNothingHasBeenWritten(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(0)

	// this is a bit of a bug...but it is not a problem.
	assert.Equal(t, []int{0}, getHistoryIds(frameLoop))
}

func getOldestFrameIds(frameLoop *FrameLoopTestClass, frames int) []int {
	ids := make([]int, frames)
	for ii := 0; ii < frames; ii++ {
		ids[ii] = getId(frameLoop.Oldest())
		frameLoop.Move()
	}
	return ids
}

func TestFrameOldest(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(1)

	assert.Equal(t, []int{1, 1, 1, 1, 1, 2, 3}, getOldestFrameIds(frameLoop, 7))

	assert.Equal(t, 8, getId(frameLoop.Current()))
	frameLoop.SetAsOldest()
	assert.Equal(t, []int{8, 8, 8, 8, 8, 9, 10}, getOldestFrameIds(frameLoop, 7))
}

func TestFrameHistoryAfterOldestSet(t *testing.T) {
	frameLoop := NewFrameLoopTestClass(8)

	frameLoop.SetAsOldest()
	assert.Equal(t, []int{8}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{8, 9}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{8, 9, 10}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{8, 9, 10, 11}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{8, 9, 10, 11, 12}, getHistoryIds(frameLoop))
	frameLoop.Move()
	assert.Equal(t, []int{9, 10, 11, 12, 13}, getHistoryIds(frameLoop))
}
