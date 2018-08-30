// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

type TestRecorder struct {
	frameIds        [200]int
	index           int
	CanRecordReturn error
}

func (tr *TestRecorder) StopRecording() error  { return nil }
func (tr *TestRecorder) StartRecording() error { return nil }
func (tr *TestRecorder) CheckCanRecord() error { return tr.CanRecordReturn }

func (tr *TestRecorder) WriteFrame(frame *lepton3.Frame) error {
	tr.frameIds[tr.index] = int(frame[0][0])
	tr.index++
	return nil
}

func (tr *TestRecorder) GetRecordedFramesIds() []int {
	return tr.frameIds[:tr.index]
}

func (tr *TestRecorder) SetCheckError(err error) {
	tr.CanRecordReturn = err
}

func DefaultTestConfig() *Config {
	config := new(Config)
	config.MinSecs = 3
	config.MaxSecs = 20
	config.PreviewSecs = 1

	config.Motion.TempThresh = 3000
	config.Motion.DeltaThresh = 50
	config.Motion.CountThresh = 3
	config.Motion.NonzeroMaxPercent = 50
	config.Motion.FrameCompareGap = 18
	config.Motion.Verbose = false
	config.Motion.TriggerFrames = 1
	config.Motion.UseOneFrameOnly = true
	config.Motion.WarmerOnly = true
	return config
}

func FramesFrom(start, end int) []int {
	slice := make([]int, end-start+1)
	for i := 0; i < end-start+1; i++ {
		slice[i] = i + start
	}
	return slice
}

func SetupTest(config *Config) (*TestRecorder, *TestFrameMaker) {
	recorder := new(TestRecorder)
	processor := NewMotionProcessor(config, new(HardwareListener), recorder)

	scenarioMaker := MakeTestFrameMaker(processor)
	return recorder, scenarioMaker
}

func TestRecorderNotTriggeredUnlessSeesMovement(t *testing.T) {
	recorder, scenarioMaker := SetupTest(DefaultTestConfig())
	scenarioMaker.AddBackgroundFrames(20)
	assert.Equal(t, []int{}, recorder.GetRecordedFramesIds())
}

func TestRecorderTriggeredAndHasPreviewAndMinNumberFrames(t *testing.T) {
	recorder, scenarioMaker := SetupTest(DefaultTestConfig())
	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(1).AddBackgroundFrames(20)
	assert.Equal(t, FramesFrom(2, 30), recorder.GetRecordedFramesIds())
}

func TestRecorderNotTriggeredUntilTriggerFramesReached(t *testing.T) {
	config := DefaultTestConfig()
	config.Motion.TriggerFrames = 3
	recorder, scenarioMaker := SetupTest(config)

	// not triggered by 2 moving frames in a row
	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(2).AddBackgroundFrames(18)
	assert.Equal(t, []int{}, recorder.GetRecordedFramesIds())

	// not triggered by 3 moving frames in a row
	scenarioMaker.AddMovingDotFrames(3).AddBackgroundFrames(40)
	assert.Equal(t, FramesFrom(24, 58), recorder.GetRecordedFramesIds())
}

func TestRecorderNotStartedIfCheckCanRecordReturnsError(t *testing.T) {
	recorder, scenarioMaker := SetupTest(DefaultTestConfig())
	recorder.SetCheckError(errors.New("Cannot record or bad things will happen"))

	// record not triggered due to error return above
	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(2).AddBackgroundFrames(18)
	assert.Equal(t, []int{}, recorder.GetRecordedFramesIds())
}
