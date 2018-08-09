// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

func defaultMotionParams() MotionConfig {
	return MotionConfig{
		TempThresh:        3000,
		DeltaThresh:       30,
		CountThresh:       8,
		NonzeroMaxPercent: 50,
		FrameCompareGap:   3,
	}
}

func makeFrame(position, background, brightSpot int) *lepton3.Frame {
	frame := new(lepton3.Frame)

	if background != 0 {
		for y := 0; y < lepton3.FrameRows; y++ {
			for x := 0; x < lepton3.FrameCols; x++ {
				frame[y][x] = uint16(background)
			}
		}
	}

	brightness16 := uint16(background + brightSpot)

	frame[position][position] = brightness16
	frame[position+1][position] = brightness16
	frame[position+2][position] = brightness16
	frame[position][position+1] = brightness16
	frame[position+1][position+1] = brightness16
	frame[position+2][position+1] = brightness16
	frame[position][position+2] = brightness16
	frame[position+1][position+2] = brightness16
	frame[position+2][position+2] = brightness16

	return frame
}

func MovingBoxDetections(detector *motionDetector, frames, background, brightSpot int) ([]int, []bool) {
	results := make([]bool, frames)
	pixels := make([]int, frames)

	for i := range results {
		results[i], pixels[i] = detector.pixelsChanged(makeFrame(10+i, background, i*100))
	}
	return pixels, results
}

func TestNoMotionDetectedUntilBufferFilledUp(t *testing.T) {
	config := defaultMotionParams()
	config.UseOneFrameOnly = true
	detector := NewMotionDetector(config)

	pixels, detecteds := MovingBoxDetections(detector, 5, 3300, 100)
	assert.Equal(t, []int{-1, -1, -1, 9, 18}, pixels)
	assert.Equal(t, []bool{false, false, false, true, true}, detecteds)
}

func TestIfUsingTwoFramesItOnlyCountsWhereBothFramesHaveChanged(t *testing.T) {
	config := defaultMotionParams()
	detector := NewMotionDetector(config)

	pixels, detecteds := MovingBoxDetections(detector, 5, 3300, 100)
	assert.Equal(t, []int{-1, -1, -1, 0, 5}, pixels)
	assert.Equal(t, []bool{false, false, false, false, false}, detecteds)

	// need to reduce the threshold to get the detector to triggeras
	// now only when both frames have changes does it trigger.
	config.CountThresh = 5
	detector = NewMotionDetector(config)

	pixels, detecteds = MovingBoxDetections(detector, 7, 3300, 100)
	assert.Equal(t, []int{-1, -1, -1, 0, 5, 9, 9}, pixels)
	assert.Equal(t, []bool{false, false, false, false, true, true, true}, detecteds)
}

func TestIfRecalculationHasOccurred(t *testing.T) {
	config := defaultMotionParams()
	detector := NewMotionDetector(config)

	pixels, detecteds := MovingBoxDetections(detector, 5, 3300, 100)
	assert.Equal(t, []int{-1, -1, -1, 0, 5}, pixels)
	assert.Equal(t, []bool{false, false, false, false, false}, detecteds)

	// need to reduce the threshold to get the detector to triggeras
	// now only when both frames have changes does it trigger.
	config.CountThresh = 5
	detector = NewMotionDetector(config)

	pixels, detecteds = MovingBoxDetections(detector, 7, 3300, 100)
	assert.Equal(t, []int{-1, -1, -1, 0, 5, 9, 9}, pixels)
	assert.Equal(t, []bool{false, false, false, false, true, true, true}, detecteds)
}
