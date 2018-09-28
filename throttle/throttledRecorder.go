package throttle

import (
	"log"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
)

// ThrottledRecorder wraps a standard recorder so that it stops recording (ie gets throttled) if requested to
// record too often.  This is desirable as the extra recordings are likely to be highly similar to the earlier recordings
// and contain no new information.  It can happen when an animal is stuck in a trap or it is very windy.
// When throttled, some recordings can still be made occasionally using the sparse recording feature.
//
// Implementation:
// The ThrottledRecorder will record as long as the mainBucket has some 'recording' tokens.  This bucket is
// filled when recorder is not recording nor throttled and empties as recordings are made.
// The sparse bucket allows occasional recordings when the device is throttled (ie not actually recording but
// detecting movement).  This bucket is completely emptied whenever a new recording starts.   It is filled whenever
// the recorder is asked to record.  This results in a new recording only after device has been throttled for a
// given time period.
type ThrottledRecorder struct {
	recorder              recorder.Recorder
	mainBucket            TokenBucket
	sparseBucket          TokenBucket
	recording             bool
	minRecordingLength    float64
	sparseRecordingLength float64
	askedToWriteFrame     bool
	throttledFrames       uint32
	frameCount            uint32
	refillRate            float64
}

func NewThrottledRecorder(baseRecorder recorder.Recorder, config *ThrottlerConfig, minSeconds int) *ThrottledRecorder {
	framesHz := uint16(lepton3.FramesHz)
	sparseFrames := config.SparseLength * framesHz
	minFrames := uint16(minSeconds) * framesHz

	if sparseFrames > 0 && sparseFrames < minFrames {
		sparseFrames = minFrames
	}

	mainBucketSize := float64(config.ThrottleAfter * framesHz)
	supBucketSize := float64(config.SparseAfter * framesHz)
	return &ThrottledRecorder{
		recorder:              baseRecorder,
		mainBucket:            TokenBucket{tokens: mainBucketSize, size: mainBucketSize},
		sparseBucket:          TokenBucket{size: supBucketSize},
		minRecordingLength:    float64(minFrames),
		sparseRecordingLength: float64(sparseFrames),
		refillRate:            config.RefillRate,
	}
}

func (throttler *ThrottledRecorder) NextFrame() {
	if throttler.askedToWriteFrame {
		throttler.mainBucket.RemoveTokens(1)
		throttler.sparseBucket.AddTokens(1)
	} else {
		throttler.mainBucket.AddTokens(throttler.refillRate)
	}
	throttler.askedToWriteFrame = false
}

func (throttler *ThrottledRecorder) CheckCanRecord() error {
	return throttler.recorder.CheckCanRecord()
}

func (throttler *ThrottledRecorder) StartRecording() error {
	if throttler.sparseBucket.IsFull() {
		log.Print("Sparse recording starting soon...")
		throttler.mainBucket.AddTokens(throttler.sparseRecordingLength)
	}

	if throttler.mainBucket.HasTokens(throttler.minRecordingLength) {
		throttler.recording = true
		throttler.sparseBucket.Empty()
		return throttler.recorder.StartRecording()
	} else {
		throttler.recording = false
		log.Print("Recording throttled")
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
		if throttler.mainBucket.HasTokens(1) {
			return throttler.recorder.WriteFrame(frame)
		} else {
			throttler.throttledFrames++
		}
	}

	return nil
}
