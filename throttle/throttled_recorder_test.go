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

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
)

type TestCamera struct {
}

func (cam *TestCamera) ResX() int {
	return 160
}
func (cam *TestCamera) ResY() int {
	return 120
}
func (cam *TestCamera) FPS() int {
	return 9
}

const (
	throttleAfter = 30 * time.Second

	minRecordingSecs   = 10
	minRecordingFrames = minRecordingSecs * lepton3.FramesHz

	minRefill = 20 * time.Second
)

var throttleFrames = int(throttleAfter.Seconds() * lepton3.FramesHz)

func newTestConfig() *config.ThermalThrottler {
	return &config.ThermalThrottler{
		Activate:   true,
		BucketSize: throttleAfter,
		MinRefill:  minRefill,
	}
}

func newTestThrottledRecorder() (*writeRecorder, *throttleListener, *ThrottledRecorder, *testClock) {
	clock := new(testClock)
	recorder := new(writeRecorder)
	listener := new(throttleListener)
	return recorder, listener, NewThrottledRecorderWithClock(recorder, newTestConfig(), minRecordingSecs, listener, clock, new(TestCamera)), clock
}

type writeRecorder struct {
	recorder.NoWriteRecorder
	writes int
}

func (rec *writeRecorder) WriteFrame(frame *cptvframe.Frame) error {
	rec.writes++
	return nil
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
	recorder.StartRecording(nil, 0)
	writeFrames(recorder, frames)
	recorder.StopRecording()
}

func writeFrames(recorder *ThrottledRecorder, frames int) {

	f := cptvframe.NewFrame(new(TestCamera))
	for i := 0; i < frames; i++ {
		recorder.WriteFrame(f)
	}
}

func TestOnlyWritesUntilBucketIsFull(t *testing.T) {
	recorder, listener, throtRecorder, _ := newTestThrottledRecorder()

	recordFrames(throtRecorder, throttleFrames+2)
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

	recordFrames(throtRecorder, throttleFrames-5)

	// only a few frames in the bucket - not enough to start another recording
	recorder.Reset()
	recordFrames(throtRecorder, 10)
	assert.Equal(t, 0, recorder.writes)
}

func TestNotRecordingFillsBucket(t *testing.T) {
	recorder, _, throtRecorder, clock := newTestThrottledRecorder()

	recordFrames(throtRecorder, throttleFrames) // empty bucket
	clock.Sleep(minRefill)                      // allow bucket to fill

	// Observe that it only filled up to the minimum size
	recorder.Reset()
	recordFrames(throtRecorder, throttleFrames)
	assert.Equal(t, minRecordingFrames, recorder.writes)
}

func TestNotifiesWhenThrottling(t *testing.T) {
	_, listener, throtRecorder, _ := newTestThrottledRecorder()

	recordFrames(throtRecorder, throttleFrames-2)
	assert.Equal(t, 0, listener.events)

	recordFrames(throtRecorder, 3)
	assert.Equal(t, 1, listener.events)
}

func TestNotifiesEvenWhenRecordingDoesntStart(t *testing.T) {
	_, listener, throtRecorder, clock := newTestThrottledRecorder()

	recordFrames(throtRecorder, throttleFrames+1)
	assert.Equal(t, 1, listener.events)

	clock.Sleep(minRefill / time.Duration(2))

	recordFrames(throtRecorder, throttleFrames)
	assert.Equal(t, 2, listener.events)
}

func TestIntraRecordingRestart(t *testing.T) {
	recorder, listener, throtRecorder, clock := newTestThrottledRecorder()

	recorder.StartRecording(nil, 0)              // start a fresh recording
	writeFrames(throtRecorder, throttleFrames+1) // Trigger throttling.
	assert.Equal(t, 1, listener.events)

	// Wait a while (recording still active) - not long enough for minimum refill.
	recorder.Reset()
	clock.Sleep(minRefill / 2)
	writeFrames(throtRecorder, 10)
	assert.Equal(t, 0, recorder.writes)

	// Wait a while long (recording still active) - should refill bucket enough.
	recorder.Reset()
	clock.Sleep(minRefill / 2)
	writeFrames(throtRecorder, 10)
	assert.Equal(t, 10, recorder.writes)

	// Throttling has only happened once (at the top).
	assert.Equal(t, 1, listener.events)
}

func TestUsingDifferentRefillRate(t *testing.T) {
	clock := new(testClock)

	config := newTestConfig()
	config.MinRefill = 60 * time.Second
	recorder := new(writeRecorder)
	throtRecorder := NewThrottledRecorderWithClock(recorder, config, minRecordingSecs, nil, clock, new(TestCamera))

	recordFrames(throtRecorder, throttleFrames) //empty bucket
	clock.Sleep(config.MinRefill)               // allow to fill

	// Observe that it only filled up to the minimum size
	recorder.Reset()
	recordFrames(throtRecorder, throttleFrames)
	assert.Equal(t, minRecordingFrames, recorder.writes)
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
