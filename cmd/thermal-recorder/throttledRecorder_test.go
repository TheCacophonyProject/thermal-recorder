package main

import (
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

const MAX_FRAMES int = 18
const MIN_FRAMES_PER_RECORDING int = 8
const SUPP_STARTFRAMES int = 99
const SUPP_LENGTH int = 9

func NewThrottledRecorder() (*CountWritesRecorder, *ThrottledRecorder) {
	recorder := new(CountWritesRecorder)

	config := &ThrottlerConfig{
		ThrottleAfter:     2,
		OccasionalAfter:   11,
		OccassionalLength: 1,
	}
	return recorder, NewRecordingThrottler(recorder, config, uint16(MIN_FRAMES_PER_RECORDING))
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
	baseRecorder.Reset()

	PlayRecordingFrames(recorder, 10)
	assert.Equal(t, 8, baseRecorder.writes)
}

func TestWillNotStartRecordingIfLessThanMinFramesToFillBucket(t *testing.T) {
	baseRecorder, recorder := NewThrottledRecorder()

	PlayRecordingFrames(recorder, 15)
	baseRecorder.Reset()
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

func TestSupplementaryRecordingWillStart(t *testing.T) {
	baseRecorder, recorder := NewThrottledRecorder()
	// fill bucket to supplementary point
	PlayRecordingFrames(recorder, 50)
	PlayRecordingFrames(recorder, 60)
	baseRecorder.Reset()

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, SUPP_LENGTH, baseRecorder.writes)
}

func TestSupplementaryRecordingCountRestartsWhenARealRecordingStarts(t *testing.T) {
	baseRecorder, recorder := NewThrottledRecorder()
	// fill bucket to supplementary point
	PlayRecordingFrames(recorder, 20)
	PlayNonRecordingFrames(recorder, int(MIN_FRAMES_PER_RECORDING+1))
	baseRecorder.Reset()

	PlayRecordingFrames(recorder, 60)
	assert.Equal(t, MIN_FRAMES_PER_RECORDING+1, baseRecorder.writes)
	baseRecorder.Reset()

	// won't start recording because supplementary count reset with start of previous recording
	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, 0, baseRecorder.writes)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, SUPP_LENGTH, baseRecorder.writes)
}

func TestCanHaveNoSupplementaryRecording(t *testing.T) {
	baseRecorder := new(CountWritesRecorder)

	config := &ThrottlerConfig{
		ThrottleAfter:     2,
		OccasionalAfter:   11,
		OccassionalLength: 0,
	}
	recorder := NewRecordingThrottler(baseRecorder, config, uint16(MIN_FRAMES_PER_RECORDING))

	PlayRecordingFrames(recorder, 50)
	PlayRecordingFrames(recorder, 60)
	baseRecorder.Reset()

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, 0, baseRecorder.writes)
}
