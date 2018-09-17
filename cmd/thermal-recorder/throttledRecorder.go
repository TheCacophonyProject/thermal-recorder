package main

import (
	"github.com/TheCacophonyProject/lepton3"
)

type ThrottledRecorder struct {
	recorder            Recorder
	mainBucket          TokenBucket
	suppBucket          TokenBucket
	recording           bool
	minRecordingLength  int32
	suppRecordingLength int32
	askedToWriteFrame   bool
}

func NewRecordingThrottler(baseRecorder Recorder, bucketSize, minFrames int, suppBucketSize, supFrames int) *ThrottledRecorder {
	if supFrames < minFrames {
		supFrames = minFrames
	}

	mainBucketSize := int32(bucketSize)
	supBucketSize := int32(suppBucketSize)
	return &ThrottledRecorder{
		recorder:            baseRecorder,
		mainBucket:          TokenBucket{tokens: mainBucketSize, size: mainBucketSize},
		suppBucket:          TokenBucket{size: supBucketSize},
		minRecordingLength:  int32(minFrames),
		suppRecordingLength: int32(supFrames),
	}
}

type TokenBucket struct {
	tokens int32
	size   int32
}

func (bucket *TokenBucket) AddTokens(newTokens int32) {
	bucket.tokens += newTokens
	if bucket.tokens > bucket.size {
		bucket.tokens = bucket.size
	}
}

func (bucket *TokenBucket) RemoveTokens(oldTokens int32) {
	if bucket.tokens >= oldTokens {
		bucket.tokens -= oldTokens
	} else {
		bucket.tokens = 0
	}
}

func (bucket TokenBucket) HasTokens(tokens int32) bool {
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
