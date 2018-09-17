package main

import (
	"github.com/TheCacophonyProject/lepton3"
)

type ThrottledRecorder struct {
	recorder            Recorder
	mainBucket          TokenBucket
	suppBucket          TokenBucket
	recording           bool
	minRecordingLength  uint32
	suppRecordingLength uint32
	askedToWriteFrame   bool
}

func NewRecordingThrottler(baseRecorder Recorder, config *ThrottlerConfig, minFrames uint16) *ThrottledRecorder {
	supFrames := config.OccassionalLength * framesHz

	if supFrames > 0 && supFrames < minFrames {
		supFrames = minFrames
	}

	mainBucketSize := uint32(config.ThrottleAfter * framesHz)
	supBucketSize := uint32(config.OccasionalAfter * framesHz)
	return &ThrottledRecorder{
		recorder:            baseRecorder,
		mainBucket:          TokenBucket{tokens: mainBucketSize, size: mainBucketSize},
		suppBucket:          TokenBucket{size: supBucketSize},
		minRecordingLength:  uint32(minFrames),
		suppRecordingLength: uint32(supFrames),
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

func (bucket TokenBucket) HasTokens(tokens uint32) bool {
	return bucket.tokens >= tokens
}

func (bucket TokenBucket) Empty() {
	bucket.tokens = 0
}

func (bucket TokenBucket) IsFull() bool {
	return bucket.HasTokens(bucket.size)
}

func (throttler *ThrottledRecorder) NewFrame() {
	if !throttler.askedToWriteFrame {
		throttler.mainBucket.AddTokens(1)
	} else {
		throttler.mainBucket.RemoveTokens(1)
		throttler.suppBucket.AddTokens(1)
	}
	throttler.askedToWriteFrame = false
}

func (throttler *ThrottledRecorder) CheckCanRecord() error {
	return throttler.recorder.CheckCanRecord()
}

func (throttler *ThrottledRecorder) StartRecording() error {
	if throttler.suppBucket.IsFull() {
		throttler.mainBucket.AddTokens(throttler.suppRecordingLength)
	}

	if throttler.mainBucket.HasTokens(throttler.minRecordingLength) {
		throttler.recording = true
		return throttler.recorder.StartRecording()
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

func (throttler *ThrottledRecorder) WriteFrame(frame *lepton3.Frame) error {
	throttler.askedToWriteFrame = true
	if throttler.mainBucket.HasTokens(1) {
		if throttler.recording {
			return throttler.recorder.WriteFrame(frame)
		}
	}

	return nil
}

func (throttler *ThrottledRecorder) IsRecording() bool {
	return throttler.recording
}
