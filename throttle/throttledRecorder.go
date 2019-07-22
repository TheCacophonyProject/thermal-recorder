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
	"log"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
)

const framesHz = uint16(lepton3.FramesHz)

// ThrottledRecorder wraps a standard recorder so that it stops recording (ie gets throttled) if requested to
// record too often.  This is desirable as the extra recordings are likely to be highly similar to the earlier recordings
// and contain no new information.  It can happen when an animal is stuck in a trap or it is very windy.

type ThrottledRecorder struct {
	recorder           recorder.Recorder
	listener           ThrottledEventListener
	bucket             TokenBucket
	recording          bool
	minRecordingLength float64
	askedToWriteFrame  bool
	throttledFrames    uint32
	frameCount         uint32
	refillRate         float64
}

type ThrottledEventListener interface {
	WhenThrottled()
}

func NewThrottledRecorder(
	baseRecorder recorder.Recorder,
	eventListener ThrottledEventListener,
	config *ThrottlerConfig,
	minSeconds int,
) *ThrottledRecorder {
	bucketSize := float64(config.ThrottleAfter * framesHz)
	return &ThrottledRecorder{
		recorder:           baseRecorder,
		listener:           eventListener,
		bucket:             TokenBucket{tokens: bucketSize, size: bucketSize},
		minRecordingLength: float64(uint16(minSeconds) * framesHz),
		refillRate:         config.RefillRate,
	}
}

func (throttler *ThrottledRecorder) NextFrame() {
	if throttler.askedToWriteFrame {
		throttler.bucket.RemoveTokens(1)
	} else {
		throttler.bucket.AddTokens(throttler.refillRate)
	}
	throttler.askedToWriteFrame = false
}

func (throttler *ThrottledRecorder) CheckCanRecord() error {
	return throttler.recorder.CheckCanRecord()
}

func (throttler *ThrottledRecorder) StartRecording() error {
	if throttler.bucket.HasTokens(throttler.minRecordingLength) {
		throttler.recording = true
		return throttler.recorder.StartRecording()
	} else {
		throttler.recording = false
		log.Print("Recording not started - currently throttled")
		if throttler.listener != nil {
			throttler.listener.WhenThrottled()
		}
		return nil
	}
}

func (throttler *ThrottledRecorder) StopRecording() error {
	if throttler.recording && throttler.throttledFrames > 0 {
		log.Printf("Stop recording; %d/%d Frames throttled", throttler.throttledFrames, throttler.frameCount)
	}
	throttler.throttledFrames = 0
	throttler.frameCount = 0

	if throttler.recording {
		throttler.recording = false
		return throttler.recorder.StopRecording()
	}
	return nil
}

func (throttler *ThrottledRecorder) WriteFrame(frame *lepton3.Frame) error {
	throttler.askedToWriteFrame = true

	if throttler.recording {
		throttler.frameCount++
		if throttler.bucket.HasTokens(1) {
			return throttler.recorder.WriteFrame(frame)
		} else {
			if throttler.throttledFrames == 0 && throttler.listener != nil {
				log.Printf("Recording throttled.")
				throttler.listener.WhenThrottled()
			}
			throttler.throttledFrames++
		}
	}

	return nil
}

// realClock implements Clock in terms of standard time functions.
type realClock struct{}

// Now implements Clock.Now by calling time.Now.
func (realClock) Now() time.Time {
	return time.Now()
}

// Now implements Clock.Sleep by calling time.Sleep.
func (realClock) Sleep(d time.Duration) {
	time.Sleep(d)
}
