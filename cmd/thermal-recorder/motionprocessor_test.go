package main

import (
	"errors"
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

type TestRecorder struct {
	frameIds         []int
	index            int
	previousFrameIds []int
	CanRecordReturn  error
}

func (tr *TestRecorder) StopRecording() error {
	tr.previousFrameIds = tr.frameIds[:tr.index]
	tr.frameIds = nil
	return nil
}

func (tr *TestRecorder) CheckCanRecord() error { return tr.CanRecordReturn }

func (tr *TestRecorder) StartRecording() error {
	tr.frameIds = make([]int, 200)
	tr.index = 0
	return nil
}

func (tr *TestRecorder) WriteFrame(frame *lepton3.Frame) error {
	tr.frameIds[tr.index] = int(frame[0][0])
	tr.index++
	return nil
}

func (tr *TestRecorder) GetRecordedFramesIds() []int {
	return tr.previousFrameIds
}

func (tr *TestRecorder) SetCheckError(err error) {
	tr.CanRecordReturn = err
}

func (tr *TestRecorder) IsRecording() bool {
	return tr.frameIds != nil
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
	config.Motion.FrameCompareGap = 12
	config.Motion.Verbose = false
	config.Motion.TriggerFrames = 1
	config.Motion.UseOneDiffOnly = true
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
	processor := NewMotionProcessor(config, nil, recorder)

	scenarioMaker := MakeTestFrameMaker(processor)
	return recorder, scenarioMaker
}

func TestRecorderNotTriggeredUnlessSeesMovement(t *testing.T) {
	recorder, scenarioMaker := SetupTest(DefaultTestConfig())
	scenarioMaker.AddBackgroundFrames(20)
	assert.False(t, recorder.IsRecording())
}

func TestRecorderTriggeredAndHasPreviewAndMinNumberFrames(t *testing.T) {
	recorder, scenarioMaker := SetupTest(DefaultTestConfig())
	scenarioMaker.AddBackgroundFrames(11).AddMovingDotFrames(1).AddBackgroundFrames(40)
	assert.Equal(t, FramesFrom(2, 37), recorder.GetRecordedFramesIds())
}

func TestRecorderNotTriggeredUntilTriggerFramesReached(t *testing.T) {
	config := DefaultTestConfig()
	config.Motion.TriggerFrames = 3
	recorder, scenarioMaker := SetupTest(config)

	// not triggered by 2 moving frames in a row
	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(2).AddBackgroundFrames(8)
	assert.False(t, recorder.IsRecording())

	// not triggered by 3 moving frames in a row
	scenarioMaker.AddMovingDotFrames(3).AddBackgroundFrames(40)
	assert.Equal(t, FramesFrom(11, 48), recorder.GetRecordedFramesIds())
}

func TestRecorderNotStartedIfCheckCanRecordReturnsError(t *testing.T) {
	recorder, scenarioMaker := SetupTest(DefaultTestConfig())
	recorder.SetCheckError(errors.New("Cannot record or bad things will happen"))

	// record not triggered due to error return above
	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(2).AddBackgroundFrames(5)
	assert.False(t, recorder.IsRecording())
}

func TestCanMakeMultipleRecordings(t *testing.T) {
	recorder, scenarioMaker := SetupTest(DefaultTestConfig())

	scenarioMaker.AddBackgroundFrames(11).AddMovingDotFrames(1).AddBackgroundFrames(39)
	assert.Equal(t, FramesFrom(2, 37), recorder.GetRecordedFramesIds())

	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(1).AddBackgroundFrames(39)
	assert.Equal(t, FramesFrom(52, 87), recorder.GetRecordedFramesIds())
}

func TestMultipleRecordingsDontRepeatAnyFrames(t *testing.T) {
	// if the tail of the previous recording comes within the preview time of the next
	// recording then only the unwritten frames are recorded.
	recorder, scenarioMaker := SetupTest(DefaultTestConfig())

	scenarioMaker.AddBackgroundFrames(11).AddMovingDotFrames(1).AddBackgroundFrames(29)
	assert.Equal(t, FramesFrom(2, 37), recorder.GetRecordedFramesIds())

	scenarioMaker.AddMovingDotFrames(1).AddBackgroundFrames(39)
	assert.Equal(t, FramesFrom(38, 67), recorder.GetRecordedFramesIds())
}
