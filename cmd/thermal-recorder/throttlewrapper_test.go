package main

import (
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

const MAX_FRAMES int16 = 20

func NewThrottledRecorder() (*CountWritesRecorder, *ThrottleWrapper) {
	recorder := new(CountWritesRecorder)
	return recorder, &ThrottleWrapper{
		recorder:   recorder,
		bucketSize: MAX_FRAMES,
		minFrames:  8,
	}
}

type CountWritesRecorder struct {
	NoWriteRecorder
	writes int
}

func (rec *CountWritesRecorder) WriteFrame(frame *lepton3.Frame) error {
	rec.writes++
	return rec.NoWriteRecorder.WriteFrame(frame)
}

func (rec *CountWritesRecorder) Reset() { rec.writes = 0 }

func PlayNonRecordingFrames(recorder *ThrottleWrapper, frames int) {
	for count := 0; count < frames; count++ {
		recorder.NewFrame()
	}
}

func PlayRecordingFrames(recorder *ThrottleWrapper, frames int) {
	testframe := new(lepton3.Frame)

	for count := 0; count < frames; count++ {
		recorder.NewFrame()
		if count == 0 {
			recorder.StartRecording()
		}
		recorder.WriteFrame(testframe)
	}
	recorder.StopRecording()
}

func TestOnlyWritesUntilBucketIsFull(t *testing.T) {

	baseRecorder, recorder := NewThrottledRecorder()

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, MAX_FRAMES, baseRecorder.writes)
}

func TestCanRecordTwice(t *testing.T) {
	baseRecorder, recorder := NewThrottledRecorder()

	PlayRecordingFrames(recorder, 10)
	assert.Equal(t, 10, baseRecorder.writes)

	PlayRecordingFrames(recorder, 10)
	assert.Equal(t, 10, baseRecorder.writes)
}

func TestWillNotStartRecordingIfLessThanMinFramesToFillBucket(t *testing.T) {
	baseRecorder, recorder := NewThrottledRecorder()

	PlayRecordingFrames(recorder, 15)
	PlayRecordingFrames(recorder, 10)

	// not frames recorded as cannot make minimum recording length
	assert.Equal(t, 0, baseRecorder.writes)
}

func TestNotRecordingIncreasesRecordingLength(t *testing.T) {
	baseRecorder, recorder := NewThrottledRecorder()
	// fill bucket
	PlayRecordingFrames(recorder, 50)
	baseRecorder.Reset()
	PlayNonRecordingFrames(recorder, 15)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, 15, baseRecorder.writes)
}
