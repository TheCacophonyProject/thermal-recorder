package throttle

import (
	"testing"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/stretchr/testify/assert"
)

const THROTTLE_FRAMES int = 27
const MIN_FRAMES_PER_RECORDING int = 9
const SPARSE_LENGTH int = 18

var countRecorder CountWritesRecorder

func NewTestThrottledRecorder() (*CountWritesRecorder, *ThrottledRecorder) {
	config := &ThrottlerConfig{
		ApplyThrottling: true,
		ThrottleAfter:   3,
		SparseAfter:     11,
		SparseLength:    2,
	}
	return &countRecorder, NewThrottledRecorder(&countRecorder, config, 1)
}

type CountWritesRecorder struct {
	recorder.NoWriteRecorder
	writes int
}

func (rec *CountWritesRecorder) WriteFrame(frame *lepton3.Frame) error {
	rec.writes++
	return rec.NoWriteRecorder.WriteFrame(frame)
}

func (rec *CountWritesRecorder) Reset() { rec.writes = 0 }

func PlayNonRecordingFrames(recorder *ThrottledRecorder, frames int) {
	for count := 0; count < frames; count++ {
		recorder.NextFrame()
	}
}

func PlayRecordingFrames(recorder *ThrottledRecorder, frames int) {
	testframe := new(lepton3.Frame)
	countRecorder.Reset()

	for count := 0; count < frames; count++ {
		recorder.NextFrame()
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
	assert.Equal(t, THROTTLE_FRAMES, baseRecorder.writes)
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

func TestSparseRecordingWillStartAndStopAgain(t *testing.T) {
	baseRecorder, recorder := NewTestThrottledRecorder()
	// add enough frames to trigger sparse recording
	PlayRecordingFrames(recorder, 50)
	PlayRecordingFrames(recorder, 60)
	assert.Equal(t, 0, baseRecorder.writes)

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, SPARSE_LENGTH, baseRecorder.writes)

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, 0, baseRecorder.writes)

	PlayRecordingFrames(recorder, 100)

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, SPARSE_LENGTH, baseRecorder.writes)
}

func TestSparseRecordingCountRestartsWhenARealRecordingStarts(t *testing.T) {
	baseRecorder, recorder := NewTestThrottledRecorder()
	// fill bucket to trigger sparse recording
	PlayRecordingFrames(recorder, THROTTLE_FRAMES)
	PlayNonRecordingFrames(recorder, int(MIN_FRAMES_PER_RECORDING+1))

	// Sparse count restarts with start of this recording
	PlayRecordingFrames(recorder, 60)
	assert.Equal(t, MIN_FRAMES_PER_RECORDING+1, baseRecorder.writes)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, 0, baseRecorder.writes)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, SPARSE_LENGTH, baseRecorder.writes)
}

func TestCanHaveNoSparseRecordings(t *testing.T) {
	baseRecorder := new(CountWritesRecorder)

	config := &ThrottlerConfig{
		ApplyThrottling: true,
		ThrottleAfter:   3,
		SparseAfter:     11,
		SparseLength:    0,
	}
	recorder := NewThrottledRecorder(baseRecorder, config, 1)

	PlayRecordingFrames(recorder, 50)
	assert.Equal(t, THROTTLE_FRAMES, baseRecorder.writes)

	PlayRecordingFrames(recorder, 60)
	baseRecorder.Reset()

	PlayRecordingFrames(recorder, 20)
	assert.Equal(t, 0, baseRecorder.writes)
}
