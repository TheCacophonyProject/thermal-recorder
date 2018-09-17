package main

import (
	"log"

	"github.com/TheCacophonyProject/lepton3"
)

type ThrottleWrapper struct {
	recorder          Recorder
	tokens            int16
	recording         bool
	bucketSize        int16
	minFrames         int16
	askedToWriteFrame bool
}

func (tw *ThrottleWrapper) NewFrame() {
	if !tw.askedToWriteFrame && tw.tokens > 0 {
		log.Printf("removing token")
		tw.tokens--
	}
	tw.askedToWriteFrame = false
}

func (tw *ThrottleWrapper) CheckCanRecord() error {
	return tw.recorder.CheckCanRecord()
}

func (tw *ThrottleWrapper) StartRecording() error {
	if tw.bucketSize-tw.tokens > tw.minFrames {
		tw.recording = true
		return tw.recorder.StartRecording()
	}
	return nil
}

func (tw *ThrottleWrapper) StopRecording() error {
	if tw.recording {
		tw.recording = false
		return tw.recorder.StopRecording()
	}
	return nil
}

func (tw *ThrottleWrapper) WriteFrame(frame *lepton3.Frame) error {
	log.Printf("Tokens in bucket %d", tw.tokens)
	tw.askedToWriteFrame = true
	if tw.tokens < tw.bucketSize {
		tw.tokens += 1
		if tw.recording {
			return tw.recorder.WriteFrame(frame)
		}
	}

	return nil
}

func (tw *ThrottleWrapper) IsRecording() bool {
	return tw.recording
}
