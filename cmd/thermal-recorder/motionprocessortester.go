// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"io"
	"log"
	"os"

	cptv "github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/lepton3"
)

type EventLoggingRecordingListener struct {
	frameCount          int
	motionDetectedCount int
	lastDetection       int
}

func (p *EventLoggingRecordingListener) MotionDetected() {
	//log.Printf("%d: Motion Detected", p.frameCount)
	if p.frameCount-p.lastDetection > 9 {
		log.Printf("%.1f: Big gap %d", float32(p.lastDetection+1)/9, p.frameCount-p.lastDetection)
	}
	p.motionDetectedCount++
	p.lastDetection = p.frameCount
}

func (p *EventLoggingRecordingListener) RecordingStarted() {
	log.Printf("%d: Recording Started", p.frameCount)
}

func (p *EventLoggingRecordingListener) RecordingEnded() {
	log.Printf("%d: Recording Ended", p.frameCount)
}

func MotionTesterProcessCPTVFile(filename string, conf *Config) {
	log.Printf("TestFile is %s, frameGap: %d", filename, conf.Motion.FrameCompareGap)

	listener := new(EventLoggingRecordingListener)

	processor := NewMotionProcessor(conf, listener)
	defer processor.Stop()

	file, reader, err := motionTesterLoadFile(filename)
	if err != nil {
		log.Printf("Could not open file %v", err)
	}
	defer file.Close()

	frame := new(lepton3.Frame)
	for {
		log.Printf("%d, Timestampe: %v", listener.frameCount, reader.Timestamp())
		if err := reader.ReadFrame(frame); err != nil {
			if err != io.EOF {
				log.Printf("Error reading file occured %v", err)
			}
			log.Printf("Last Frame gap %d", listener.frameCount-listener.lastDetection)
			log.Printf("Motion detected frames %d out of frames %d (%d)", listener.motionDetectedCount, listener.frameCount, listener.frameCount-conf.Motion.FrameCompareGap)
			return
		}
		processor.processFrame(frame)
		listener.frameCount++
	}
}

func motionTesterLoadFile(filename string) (*os.File, *cptv.Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, nil, err
	}
	reader, err := cptv.NewReader(file)
	return file, reader, err
}
