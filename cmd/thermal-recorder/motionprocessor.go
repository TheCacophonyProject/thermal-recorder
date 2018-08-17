// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	cptv "github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/lepton3"
)

func NewMotionProcessor(conf *Config, listener RecordingListener) *MotionProcessor {
	return &MotionProcessor{
		minFrames:      conf.MinSecs * framesHz,
		maxFrames:      conf.MaxSecs * framesHz,
		motionDetector: NewMotionDetector(conf.Motion),
		frameLoop:      NewFrameLoop(conf.PreviewSecs * framesHz),
		isRecording:    false,
		window:         *NewWindow(conf.WindowStart, conf.WindowEnd),
		listener:       listener,
		conf:           conf,
		DoRecord:       true,
		triggerFrames:  conf.Motion.TriggerFrames,
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
	writer         *cptv.FileWriter
	window         Window
	conf           *Config
	listener       RecordingListener
	DoRecord       bool
	triggerFrames  int
	triggered      int
}

type RecordingListener interface {
	MotionDetected()
	RecordingStarted()
	RecordingEnded()
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

		if mp.isRecording {
			// increase the length of recording
			mp.writeUntil = min(mp.framesWritten+mp.minFrames, mp.maxFrames)
		} else if mp.triggered+1 < mp.triggerFrames {
			// Only start recording after n (triggerFrames) consecutive frames with motion detected.
			mp.triggered++
		} else if err := mp.canStartWriting(); err != nil {
			mp.SometimesWriteError("Recording not started", err)
		} else if err := mp.startRecording(); err != nil {
			mp.SometimesWriteError("Can't start recording file", err)
		} else {
			mp.writeUntil = mp.minFrames
		}
	} else {
		mp.triggered = 0
	}

	// If recording, write the frame.
	if mp.isRecording {
		if mp.DoRecord {
			err := mp.writer.WriteFrame(frame)
			if err != nil {
				log.Printf("Failed to write to CPTV file %v", err)
			}
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
	//frame.Copy(srcFrame)
	CopyFrames(frame, srcFrame)

	mp.internalProcess(frame)
}

func (mp *MotionProcessor) canStartWriting() error {
	if !mp.window.Active() {
		return errors.New("motion detected but outside of recording window")
	} else if enoughSpace, err := checkDiskSpace(mp.conf.MinDiskSpace, mp.conf.OutputDir); err != nil {
		return fmt.Errorf("Problem with disk space: %v", err)
	} else if !enoughSpace {
		return errors.New("motion detected but not enough free disk space to start recording")
	}
	return nil
}

func (mp *MotionProcessor) SometimesWriteError(task string, err error) {
	shouldLogMotion := (mp.lastLogFrame == 0) //|| (mp.totalFrames >= mp.lastLogFrame+(10*framesHz))
	if shouldLogMotion {
		log.Printf("%s (%d): %v", task, mp.totalFrames, err)
		mp.lastLogFrame = mp.totalFrames
	}
}

func (mp *MotionProcessor) startRecording() error {
	var err error

	if mp.DoRecord {
		filename := filepath.Join(mp.conf.OutputDir, newRecordingTempName())
		log.Printf("recording started: %s", filename)

		if mp.writer, err = cptv.NewFileWriter(filename); err != nil {
			return err
		}
	}

	mp.isRecording = true
	mp.listener.RecordingStarted()

	if mp.DoRecord {
		if mp.writer.WriteHeader(mp.conf.DeviceName); err != nil {
			return err
		}

		err = mp.writeInitialFramesToFile(mp.writer)
		return err
	}

	return nil
}

func (mp *MotionProcessor) stopRecording() error {
	mp.listener.RecordingEnded()
	var returnErr error

	if mp.DoRecord {

		mp.writer.Close()

		finalName, err := renameTempRecording(mp.writer.Name())
		returnErr = err
		log.Printf("recording stopped: %s\n", finalName)
	}

	mp.writer = nil
	mp.framesWritten = 0
	mp.writeUntil = 0
	mp.isRecording = false
	mp.triggered = 0
	// if it starts recording again very quickly it won't write the same frames again
	mp.frameLoop.SetAsOldest()

	return returnErr
}

func (mp *MotionProcessor) Stop() {
	if mp.writer != nil {
		mp.writer.Close()
		os.Remove(mp.writer.Name())
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func newRecordingTempName() string {
	return time.Now().Format("20060102.150405.000." + cptvTempExt)
}

func (mp *MotionProcessor) writeInitialFramesToFile(writer *cptv.FileWriter) error {
	frames := mp.frameLoop.GetHistory()
	var frame *lepton3.Frame
	ii := 0

	// it never writes the current frame as this will be written as part of the program!!
	for ii < len(frames)-1 {
		frame = frames[ii]
		if err := writer.WriteFrame(frame); err != nil {
			return err
		}
		ii++
	}

	return nil
}

// Copy sets current frame as other frame
func CopyFrames(dest *lepton3.Frame, orig *lepton3.Frame) {
	for y := 0; y < lepton3.FrameRows; y++ {
		copy(dest[y][:], orig[y][:])
	}
}
