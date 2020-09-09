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

	"github.com/juju/ratelimit"

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
)

func NewThrottledRecorder(
	baseRecorder recorder.Recorder,
	config *config.ThermalThrottler,
	minSeconds int,
	eventListener ThrottledEventListener, camera cptvframe.CameraSpec,
) *ThrottledRecorder {
	return NewThrottledRecorderWithClock(
		baseRecorder,
		config,
		minSeconds,
		eventListener,
		new(realClock), camera,
	)
}

func NewThrottledRecorderWithClock(
	baseRecorder recorder.Recorder,
	config *config.ThermalThrottler,
	minSeconds int,
	listener ThrottledEventListener,
	clock ratelimit.Clock, camera cptvframe.CameraSpec,
) *ThrottledRecorder {
	// The token bucket tracks the number of *frames* available for recording.
	bucketFrames := int64(config.BucketSize.Seconds()) * int64(camera.FPS())
	minFrames := int64(minSeconds * camera.FPS())
	refillRate := float64(minFrames) / config.MinRefill.Seconds()

	if minFrames > bucketFrames {
		log.Println("minimum recording length is greater than throttle bucket - recording will not be possible!")
	}

	bucket := ratelimit.NewBucketWithRateAndClock(refillRate, bucketFrames, clock)

	if listener == nil {
		listener = new(nullListener)
	}

	return &ThrottledRecorder{
		recorder:           baseRecorder,
		listener:           listener,
		bucket:             bucket,
		minRecordingLength: minFrames,
	}
}

// ThrottledRecorder wraps a standard recorder so that it stops
// recording (ie gets throttled) if requested to record too often.
// This is desirable as the extra recordings are likely to be highly
// similar to the earlier recordings and contain no new information.
// It can happen when an animal is stuck in a trap or it is very
// windy.
type ThrottledRecorder struct {
	recorder           recorder.Recorder
	listener           ThrottledEventListener
	bucket             *ratelimit.Bucket
	recording          bool
	minRecordingLength int64
}

type ThrottledEventListener interface {
	WhenThrottled()
}

type nullListener struct{}

func (lis *nullListener) WhenThrottled() {}

func (throttler *ThrottledRecorder) CheckCanRecord() error {
	return throttler.recorder.CheckCanRecord()
}

func (throttler *ThrottledRecorder) StartRecording(tempThresh uint16) error {
	if err := throttler.maybeStartRecording(tempThresh); err != nil {
		return err
	}
	if !throttler.recording {
		log.Print("recording not started due to throttling")
		throttler.listener.WhenThrottled()
	}
	return nil
}

func (throttler *ThrottledRecorder) StopRecording() error {
	if throttler.recording {
		throttler.recording = false
		return throttler.recorder.StopRecording()
	}
	return nil
}

func (throttler *ThrottledRecorder) WriteFrame(frame *cptvframe.Frame, tempThresh uint16) error {
	if !throttler.recording {
		if err := throttler.maybeStartRecording(tempThresh); err != nil {
			return err
		}
		if !throttler.recording {
			return nil
		}
	}

	if throttler.bucket.TakeAvailable(1) > 0 {
		return throttler.recorder.WriteFrame(frame, tempThresh)
	}

	log.Print("recording throttled")
	throttler.listener.WhenThrottled()
	return throttler.StopRecording()
}

func (throttler *ThrottledRecorder) maybeStartRecording(tempThresh uint16) error {
	if throttler.bucket.Available() >= throttler.minRecordingLength {
		if err := throttler.recorder.StartRecording(tempThresh); err != nil {
			return err
		}
		throttler.recording = true
	}
	return nil
}

// realClock implements ratelimit.Clock in terms of standard time functions.
type realClock struct{}

// Now implements Clock.Now by calling time.Now.
func (realClock) Now() time.Time {
	return time.Now()
}

// Now implements Clock.Sleep by calling time.Sleep.
func (realClock) Sleep(d time.Duration) {
	time.Sleep(d)
}
