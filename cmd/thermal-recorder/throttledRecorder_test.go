package main

import (
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

const MAX_FRAMES int = 27
const MIN_FRAMES_PER_RECORDING int = 9
const OCC_STARTFRAMES int = 99
const OCC_LENGTH int = 18

var baseRecorder CountWritesRecorder = CountWritesRecorder{}

func NewTestThrottledRecorder() (*CountWritesRecorder, *ThrottledRecorder) {
	config := &ThrottlerConfig{
		ThrottleAfter:     3,
		OccasionalAfter:   11,
		OccassionalLength: 2,
	}
	return &baseRecorder, NewThrottledRecorder(&baseRecorder, config, 1)
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

func PlayNonRecordingFrames(recorder *ThrottledRecorder, frames int) {
	for count := 0; count < frames; count++ {
		recorder.NewFrame()
	}
}

func PlayRecordingFrames(recorder *ThrottledRecorder, frames int) {
	testframe := new(lepton3.Frame)
	baseRecorder.Reset()

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

	baseRecorder, recorder := NewTestThrottledRecorder()

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, MAX_FRAMES, baseRecorder.writes)
}

func TestCanRecordTwice(t *testing.T) {
	baseRecorder, recorder := NewTestThrottledRecorder()

	PlayRecordingFrames(recorder, 10)
	assert.Equal(t, 10, baseRecorder.writes)

	PlayRecordingFrames(recorder, 10)
	assert.Equal(t, 10, baseRecorder.writes)
}

func TestWillNotStartRecordingIfLessThanMinFramesToFillBucket(t *testing.T) {
	baseRecorder, recorder := NewTestThrottledRecorder()

	PlayRecordingFrames(recorder, 20)

	// only 7 frames in the buffer - not enough to start another recording
	PlayRecordingFrames(recorder, 10)
	assert.Equal(t, 0, baseRecorder.writes)
}

func TestNotRecordingIncreasesRecordingLength(t *testing.T) {
	baseRecorder, recorder := NewTestThrottledRecorder()
	// fill bucket
	PlayRecordingFrames(recorder, 50)
	PlayNonRecordingFrames(recorder, 15)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, 15, baseRecorder.writes)
}

func TestOccasionalRecordingWillStartAndStopAgain(t *testing.T) {
	baseRecorder, recorder := NewTestThrottledRecorder()
	// fill bucket to supplementary point
	PlayRecordingFrames(recorder, 50)
	PlayRecordingFrames(recorder, 60)
	assert.Equal(t, 0, baseRecorder.writes)

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, OCC_LENGTH, baseRecorder.writes)

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, 0, baseRecorder.writes)

	PlayRecordingFrames(recorder, 100)

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, OCC_LENGTH, baseRecorder.writes)
}

func TestSupplementaryRecordingCountRestartsWhenARealRecordingStarts(t *testing.T) {
	baseRecorder, recorder := NewTestThrottledRecorder()
	// fill bucket to supplementary point
	PlayRecordingFrames(recorder, MAX_FRAMES)
	PlayNonRecordingFrames(recorder, int(MIN_FRAMES_PER_RECORDING+1))

	// Occassional count restarts with start of this recording
	PlayRecordingFrames(recorder, 60)
	assert.Equal(t, MIN_FRAMES_PER_RECORDING+1, baseRecorder.writes)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, 0, baseRecorder.writes)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, OCC_LENGTH, baseRecorder.writes)
}

func TestCanHaveNoSupplementaryRecording(t *testing.T) {
	baseRecorder := new(CountWritesRecorder)

	config := &ThrottlerConfig{
		ThrottleAfter:     3,
		OccasionalAfter:   11,
		OccassionalLength: 0,
	}
	recorder := NewThrottledRecorder(baseRecorder, config, 1)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, MAX_FRAMES, baseRecorder.writes)

	PlayRecordingFrames(recorder, 60)
	baseRecorder.Reset()

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, 0, baseRecorder.writes)
}
