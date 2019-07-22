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
	"time"

	"github.com/juju/ratelimit"
	"github.com/stretchr/testify/assert"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
)

const throttleFrames int = 27
const minFramesPerRecording int = 9

func newTestConfig() *ThrottlerConfig {
	return &ThrottlerConfig{
		ApplyThrottling: true,
		ThrottleAfter:   3,
		RefillRate:      1.0,
	}
}

func newTestThrottledRecorder() (*writeRecorder, *throttleListener, *ThrottledRecorder, *testClock) {
	clock := new(testClock)
	recorder := new(writeRecorder)
	listener := new(throttleListener)
	return recorder, listener, NewThrottledRecorderWithClock(recorder, newTestConfig(), 1, listener, clock), clock
}

type writeRecorder struct {
	recorder.NoWriteRecorder
	writes int
}

func (rec *writeRecorder) WriteFrame(frame *lepton3.Frame) error {
	rec.writes++
	return rec.NoWriteRecorder.WriteFrame(frame)
}

func (rec *writeRecorder) Reset() {
	rec.writes = 0
}

type throttleListener struct {
	events int
}

func (tc *throttleListener) WhenThrottled() {
	tc.events++
}

func recordFrames(recorder *ThrottledRecorder, frames int) {
	recorder.StartRecording()

	testframe := new(lepton3.Frame)
	for i := 0; i < frames; i++ {
		recorder.WriteFrame(testframe)
	}
	recorder.StopRecording()
}

func TestOnlyWritesUntilBucketIsFull(t *testing.T) {
	recorder, listener, throtRecorder, _ := newTestThrottledRecorder()

	recordFrames(throtRecorder, 50)
	assert.Equal(t, throttleFrames, recorder.writes)
	assert.Equal(t, 1, listener.events)
}

func TestCanRecordTwiceWithoutThrottling(t *testing.T) {
	recorder, _, throtRecorder, _ := newTestThrottledRecorder()

	recordFrames(throtRecorder, 10)
	assert.Equal(t, 10, recorder.writes)

	recordFrames(throtRecorder, 10)
	assert.Equal(t, 20, recorder.writes)
}

func TestWillNotStartRecordingIfLessThanMinFramesToFillBucket(t *testing.T) {
	recorder, _, throtRecorder, _ := newTestThrottledRecorder()

	recordFrames(throtRecorder, 20)

	// only 7 frames in the bucket - not enough to start another recording
	recorder.Reset()
	recordFrames(throtRecorder, 10)
	assert.Equal(t, 0, recorder.writes)
}

func TestNotRecordingFillsBucket(t *testing.T) {
	recorder, _, throtRecorder, clock := newTestThrottledRecorder()

	// empty bucket
	recordFrames(throtRecorder, 50)

	// allow it fill a bit
	clock.Sleep(15 * time.Second)

	// Observe that it only filled a bit
	recorder.Reset()
	recordFrames(throtRecorder, 50)
	assert.Equal(t, 15, recorder.writes)
}

func TestUsingDifferentRefillRates(t *testing.T) {
	clock := new(testClock)

	config := newTestConfig()
	config.RefillRate = 3

	recorder := new(writeRecorder)
	throtRecorder := NewThrottledRecorderWithClock(recorder, config, 1, nil, clock)
	recordFrames(throtRecorder, throttleFrames)
	clock.Sleep(5 * time.Second)
	recorder.Reset()
	recordFrames(throtRecorder, 60)
	assert.Equal(t, 15, recorder.writes)

	recorder.Reset()
	config.RefillRate = .3
	throtRecorder = NewThrottledRecorderWithClock(recorder, config, 1, nil, clock)
	recordFrames(throtRecorder, throttleFrames)
	clock.Sleep(31 * time.Second)
	recorder.Reset()
	recordFrames(throtRecorder, 60)
	assert.Equal(t, 9, recorder.writes)
}

func TestNotifiesWhenThrottling(t *testing.T) {
	_, listener, throtRecorder, _ := newTestThrottledRecorder()

	recordFrames(throtRecorder, 10)
	assert.Equal(t, 0, listener.events)

	recordFrames(throtRecorder, 40)
	assert.Equal(t, 1, listener.events)
}

func TestNotifiesEvenWhenRecordingDoesntStart(t *testing.T) {
	_, listener, throtRecorder, clock := newTestThrottledRecorder()

	recordFrames(throtRecorder, 50)
	assert.Equal(t, 1, listener.events)

	clock.Sleep(time.Duration((minFramesPerRecording-2)*lepton3.FramesHz) * time.Second)

	recordFrames(throtRecorder, 50)
	assert.Equal(t, 2, listener.events)
}

var _ ratelimit.Clock = new(realClock)
var _ ratelimit.Clock = new(testClock)

// testClock implements a fake ratelimit.Clock for testing.
type testClock struct {
	now time.Time
}

// Now implements Clock.Now by calling time.Now.
func (c *testClock) Now() time.Time {
	return c.now
}

// Now implements Clock.Sleep by calling time.Sleep.
func (c *testClock) Sleep(d time.Duration) {
	c.now = c.now.Add(d)
}
