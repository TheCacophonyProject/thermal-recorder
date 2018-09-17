package main

import (
	"log"

	"github.com/TheCacophonyProject/lepton3"
)

type ThrottledRecorder struct {
	recorder              Recorder
	mainBucket            TokenBucket
	sparseBucket          TokenBucket
	recording             bool
	minRecordingLength    uint32
	sparseRecordingLength uint32
	askedToWriteFrame     bool
	throttledFrames       uint32
	frameCount            uint32
}

func NewThrottledRecorder(baseRecorder Recorder, config *ThrottlerConfig, minSeconds int) *ThrottledRecorder {
	sparseFrames := config.SparseLength * framesHz
	minFrames := uint16(minSeconds) * framesHz

	if sparseFrames > 0 && sparseFrames < minFrames {
		sparseFrames = minFrames
	}

	mainBucketSize := uint32(config.ThrottleAfter * framesHz)
	supBucketSize := uint32(config.SparseAfter * framesHz)
	return &ThrottledRecorder{
		recorder:              baseRecorder,
		mainBucket:            TokenBucket{tokens: mainBucketSize, size: mainBucketSize},
		sparseBucket:          TokenBucket{size: supBucketSize},
		minRecordingLength:    uint32(minFrames),
		sparseRecordingLength: uint32(sparseFrames),
	}
}

type TokenBucket struct {
	tokens uint32
	size   uint32
}

func (bucket *TokenBucket) AddTokens(newTokens uint32) {
	bucket.tokens += newTokens
	if bucket.tokens > bucket.size {
		bucket.tokens = bucket.size
	}
}

func (bucket *TokenBucket) RemoveTokens(oldTokens uint32) {
	if bucket.tokens >= oldTokens {
		bucket.tokens -= oldTokens
	} else {
		bucket.tokens = 0
	}
}

func (bucket *TokenBucket) HasTokens(tokens uint32) bool {
	return bucket.tokens >= tokens
}

func (bucket *TokenBucket) Empty() {
	bucket.tokens = 0
}

func (bucket *TokenBucket) IsFull() bool {
	return bucket.HasTokens(bucket.size)
}

func (throttler *ThrottledRecorder) NewFrame() {
	if !throttler.askedToWriteFrame {
		throttler.mainBucket.AddTokens(1)
	} else {
		throttler.mainBucket.RemoveTokens(1)
		throttler.sparseBucket.AddTokens(1)
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

func (throttler *ThrottledRecorder) IsRecording() bool {
	return throttler.recording
}
