// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2018, The Cacophony Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package throttle

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
)

const THROTTLE_FRAMES int = 27
const MIN_FRAMES_PER_RECORDING int = 9
const SPARSE_LENGTH int = 18

var countRecorder CountWritesRecorder

func DefaultTestThrottleConfig() *ThrottlerConfig {
	return &ThrottlerConfig{
		ApplyThrottling: true,
		ThrottleAfter:   3,
		SparseAfter:     11,
		SparseLength:    2,
		RefillRate:      1.0,
	}
}

func NewTestThrottledRecorder() (*CountWritesRecorder, *ThrottledRecorder) {
	return &countRecorder, NewThrottledRecorder(&countRecorder, DefaultTestThrottleConfig(), 1)
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

func TestUsingDifferentRefillRates(t *testing.T) {
	config := DefaultTestThrottleConfig()
	config.RefillRate = 3

	recorder := NewThrottledRecorder(&countRecorder, config, 1)
	// fill bucket to trigger sparse recording
	PlayRecordingFrames(recorder, THROTTLE_FRAMES)
	PlayNonRecordingFrames(recorder, 5)

	// Sparse count restarts with start of this recording
	PlayRecordingFrames(recorder, 60)
	assert.Equal(t, 15, countRecorder.writes)

	config.RefillRate = .3
	recorder = NewThrottledRecorder(&countRecorder, config, 1)
	// fill bucket to trigger sparse recording
	PlayRecordingFrames(recorder, THROTTLE_FRAMES)
	PlayNonRecordingFrames(recorder, 31)

	// Sparse count restarts with start of this recording
	PlayRecordingFrames(recorder, 60)
	assert.Equal(t, 9, countRecorder.writes)
}
