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
	gaps                int
	frameCount          int
	motionDetectedCount int
	lastDetection       int
	verbose             bool
	recordingStarted    int
}

func (p *EventLoggingRecordingListener) MotionDetected() {
	if p.verbose {
		log.Printf("%d: Motion Detected", p.frameCount)
	}
	if p.frameCount-p.lastDetection > 18 {
		if p.verbose {
			log.Printf("%.1f: Big gap %d", float32(p.lastDetection+1)/9, p.frameCount-p.lastDetection)
		}
		p.gaps++
	}
	p.motionDetectedCount++
	p.lastDetection = p.frameCount
}

func (p *EventLoggingRecordingListener) RecordingStarted() {
	if p.verbose {
		log.Printf("%d: Recording Started", p.frameCount)
	}
	p.recordingStarted = p.frameCount
}

func (p *EventLoggingRecordingListener) RecordingEnded() {
	if p.verbose {
		log.Printf("%d: Recording Ended", p.frameCount)
	}
}

type NoWriteRecorder struct {
}

func (*NoWriteRecorder) StopRecording() error            { return nil }
func (*NoWriteRecorder) StartRecording() error           { return nil }
func (*NoWriteRecorder) WriteFrame(*lepton3.Frame) error { return nil }
func (*NoWriteRecorder) CheckCanRecord() error           { return nil }

func MotionTesterProcessMultipleCptvFiles(conf *Config) {
	path := "/Users/clare/cacophony/data/cptvfiles/"

	MotionTesterProcessCPTVFile(path+"rat.cptv", conf)
	MotionTesterProcessCPTVFile(path+"rat02.cptv", conf)
	MotionTesterProcessCPTVFile(path+"noise_01.cptv", conf)
	MotionTesterProcessCPTVFile(path+"noise_02.cptv", conf)
	MotionTesterProcessCPTVFile(path+"noise_03.cptv", conf)
	MotionTesterProcessCPTVFile(path+"noise_05.cptv", conf)
	MotionTesterProcessCPTVFile(path+"skyline.cptv", conf)
	MotionTesterProcessCPTVFile(path+"20180814-182224.cptv", conf)
	MotionTesterProcessCPTVFile(path+"20180814-153539.cptv", conf)
	MotionTesterProcessCPTVFile(path+"20180814-153527.cptv", conf)
}

func MotionTesterProcessCPTVFile(filename string, conf *Config) {
	verbose := conf.Motion.Verbose
	if verbose {
		log.Printf("TestFile is %s", filename)
	}

	listener := new(EventLoggingRecordingListener)
	listener.verbose = verbose
	listener.recordingStarted = -1

	recorder := new(NoWriteRecorder)

	processor := NewMotionProcessor(conf, listener, recorder)

	file, reader, err := motionTesterLoadFile(filename)
	if err != nil {
		log.Printf("Could not open file %v", err)
	}
	defer file.Close()

	frame := new(lepton3.Frame)
	for {
		// log.Printf("%d, Timestampe: %v", listener.frameCount, reader.Timestamp())
		if err := reader.ReadFrame(frame); err != nil {
			if verbose {
				if err != io.EOF {
					log.Printf("Error reading file occured %v", err)
				}
				log.Printf("Last Frame gap %d", listener.frameCount-listener.lastDetection)
				log.Printf("Motion detected frames %d out of frames %d (%d)", listener.motionDetectedCount, listener.frameCount, listener.gaps)
			} else {
				log.Printf("%s: %d out of frames %d (last: %d, gaps: %d) recorded from (%d) ", filename, listener.motionDetectedCount, listener.frameCount, listener.lastDetection, listener.gaps, listener.recordingStarted)
			}
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
