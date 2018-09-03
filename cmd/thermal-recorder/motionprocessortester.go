// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	cptv "github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/lepton3"
)

type EventLoggingRecordingListener struct {
	config              *Config
	gaps                int
	frameCount          int
	motionDetectedCount int
	lastDetection       int
	verbose             bool
	recordingStarted    string
	motionDetected      string
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
	p.recordingStarted += fmt.Sprintf("(%d:", p.frameCount)
	p.motionDetected += fmt.Sprintf("(%d:", p.frameCount-p.config.Motion.TriggerFrames+1)
}

func (p *EventLoggingRecordingListener) RecordingEnded() {
	if p.verbose {
		log.Printf("%d: Recording Ended", p.frameCount)
	}
	p.recordingStarted += fmt.Sprintf("%d)", p.frameCount)
	p.motionDetected += fmt.Sprintf("%d)", p.frameCount-p.config.MinSecs*9)
}

type NoWriteRecorder struct {
}

func (*NoWriteRecorder) StopRecording() error            { return nil }
func (*NoWriteRecorder) StartRecording() error           { return nil }
func (*NoWriteRecorder) WriteFrame(*lepton3.Frame) error { return nil }
func (*NoWriteRecorder) CheckCanRecord() error           { return nil }

type CPTVPlaybackTester struct {
	config   *Config
	basePath string
	results  []string
}

func NewCPTVPlaybackTester(conf *Config) *CPTVPlaybackTester {
	return &CPTVPlaybackTester{
		config:  conf,
		results: make([]string, 0, 100),
	}
}

func (cpt *CPTVPlaybackTester) processIfCPTVFile(path string, info os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".cptv") {
		log.Printf("Testing  %s", path)
		newResult := cpt.Detect(path)
		cpt.results = append(cpt.results, cpt.makeResultSummary(path, newResult))
	}
	return nil
}

func (cpt *CPTVPlaybackTester) makeResultSummary(filename string, listener *EventLoggingRecordingListener) string {
	shortName := filename[len(cpt.basePath)+1:]
	if listener.motionDetected == "" {
		listener.motionDetected = "None"
	}

	if listener.recordingStarted == "" {
		listener.recordingStarted = "None"
	}

	details := fmt.Sprintf("%-20s Detected: %-16s Recorded: %-16s Motion frames: %d/%d", shortName, listener.motionDetected, listener.recordingStarted, listener.motionDetectedCount, listener.frameCount)
	log.Print(details)
	return details
}

func (cpt *CPTVPlaybackTester) TestAllCPTVFiles(dir string) []string {
	cpt.basePath = dir
	log.Printf("Looking for CPTV files in %s", cpt.basePath)
	filepath.Walk(cpt.basePath, cpt.processIfCPTVFile)
	return cpt.results
}

func (cpt *CPTVPlaybackTester) Detect(filename string) *EventLoggingRecordingListener {
	verbose := cpt.config.Motion.Verbose
	if verbose {
		log.Printf("TestFile is %s", filename)
	}

	listener := new(EventLoggingRecordingListener)
	listener.config = cpt.config
	listener.verbose = verbose
	listener.recordingStarted = ""

	recorder := new(NoWriteRecorder)

	processor := NewMotionProcessor(cpt.config, listener, recorder)

	file, reader, err := motionTesterLoadFile(filename)
	if err != nil {
		log.Printf("Could not open file %v", err)
	}
	defer file.Close()

	frame := new(lepton3.Frame)
	for {
		if err := reader.ReadFrame(frame); err != nil {
			if verbose {
				if err != io.EOF {
					log.Printf("Error reading file occured %v", err)
				}
				log.Printf("Last Frame gap %d", listener.frameCount-listener.lastDetection)
				log.Printf("Motion detected frames %d out of frames %d (%d)", listener.motionDetectedCount, listener.frameCount, listener.gaps)
			}
			return listener
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
