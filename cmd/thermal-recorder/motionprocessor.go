// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"log"

	"github.com/TheCacophonyProject/lepton3"
)

func NewMotionProcessor(conf *Config, listener RecordingListener, recorder Recorder) *MotionProcessor {
	return &MotionProcessor{
		minFrames:      conf.MinSecs * framesHz,
		maxFrames:      conf.MaxSecs * framesHz,
		motionDetector: NewMotionDetector(conf.Motion),
		frameLoop:      NewFrameLoop(conf.PreviewSecs*framesHz + conf.Motion.TriggerFrames),
		isRecording:    false,
		window:         *NewWindow(conf.WindowStart, conf.WindowEnd),
		listener:       listener,
		conf:           conf,
		triggerFrames:  conf.Motion.TriggerFrames,
		recorder:       recorder,
	}
}

type MotionProcessor struct {
	minFrames      int
	maxFrames      int
	framesWritten  int
	motionDetector *motionDetector
	frameLoop      *FrameLoop
	isRecording    bool
	totalFrames    int
	writeUntil     int
	lastLogFrame   int
	window         Window
	conf           *Config
	listener       RecordingListener
	triggerFrames  int
	triggered      int
	recorder       Recorder
}

type RecordingListener interface {
	MotionDetected()
	RecordingStarted()
	RecordingEnded()
}

type Recorder interface {
	StopRecording() error
	StartRecording() error
	WriteFrame(*lepton3.Frame) error
	CheckCanRecord() error
}

func (mp *MotionProcessor) Process(rawFrame *lepton3.RawFrame) {
	frame := mp.frameLoop.Current()
	rawFrame.ToFrame(frame)

	mp.internalProcess(frame)
}

func (mp *MotionProcessor) internalProcess(frame *lepton3.Frame) {
	mp.totalFrames++

	if mp.motionDetector.Detect(frame) {
		mp.listener.MotionDetected()
		mp.triggered++

		if mp.isRecording {
			// increase the length of recording
			mp.writeUntil = min(mp.framesWritten+mp.minFrames, mp.maxFrames)
		} else if mp.triggered < mp.triggerFrames {
			// Only start recording after n (triggerFrames) consecutive frames with motion detected.
		} else if err := mp.canStartWriting(); err != nil {
			mp.occasionallyWriteError("Recording not started", err)
		} else if err := mp.startRecording(); err != nil {
			mp.occasionallyWriteError("Can't start recording file", err)
		} else {
			mp.writeUntil = mp.minFrames
		}
	} else {
		mp.triggered = 0
	}

	// If recording, write the frame.
	if mp.isRecording {
		err := mp.recorder.WriteFrame(frame)
		if err != nil {
			log.Printf("Failed to write to CPTV file %v", err)
		}
		mp.framesWritten++
	}

	mp.frameLoop.Move()

	if mp.isRecording && mp.framesWritten >= mp.writeUntil {
		err := mp.stopRecording()
		if err != nil {
			log.Printf("Failed to stop recording CPTV file %v", err)
		}
	}
}

func (mp *MotionProcessor) processFrame(srcFrame *lepton3.Frame) {

	frame := mp.frameLoop.Current()
	frame.Copy(srcFrame)

	mp.internalProcess(frame)
}

func (mp *MotionProcessor) canStartWriting() error {
	if !mp.window.Active() {
		return errors.New("motion detected but outside of recording window")
	} else {
		return mp.recorder.CheckCanRecord()
	}
}

func (mp *MotionProcessor) occasionallyWriteError(task string, err error) {
	shouldLogMotion := (mp.lastLogFrame == 0) //|| (mp.totalFrames >= mp.lastLogFrame+(10*framesHz))
	if shouldLogMotion {
		log.Printf("%s (%d): %v", task, mp.totalFrames, err)
		mp.lastLogFrame = mp.totalFrames
	}
}

func (mp *MotionProcessor) startRecording() error {

	var err error

	if err = mp.recorder.StartRecording(); err != nil {
		return err
	}

	mp.isRecording = true
	mp.listener.RecordingStarted()

	err = mp.recordPreTriggerFrames()
	return err
}

func (mp *MotionProcessor) stopRecording() error {
	mp.listener.RecordingEnded()
	err := mp.recorder.StopRecording()

	mp.framesWritten = 0
	mp.writeUntil = 0
	mp.isRecording = false
	mp.triggered = 0
	// if it starts recording again very quickly it won't write the same frames again
	mp.frameLoop.SetAsOldest()

	return err
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func (mp *MotionProcessor) recordPreTriggerFrames() error {
	frames := mp.frameLoop.GetHistory()
	var frame *lepton3.Frame
	ii := 0

	// it never writes the current frame as this will be written later
	for ii < len(frames)-1 {
		frame = frames[ii]
		if err := mp.recorder.WriteFrame(frame); err != nil {
			return err
		}
		ii++
	}

	return nil
}
