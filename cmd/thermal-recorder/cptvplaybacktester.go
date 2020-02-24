// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2018, The Cacophony Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	cptv "github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/lepton3"

	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
)

type EventLoggingRecordingListener struct {
	config               *Config
	frameCount           int
	motionDetectedCount  int
	lastDetection        int
	verbose              bool
	recordedFrames       string
	motionDetectedFrames string
	framesHz             int
}

func (p *EventLoggingRecordingListener) MotionDetected() {
	if p.verbose {
		log.Printf("%d: Motion Detected", p.frameCount)
	}
	p.motionDetectedCount++
	p.lastDetection = p.frameCount
}

func (p *EventLoggingRecordingListener) RecordingStarted() {
	if p.verbose {
		log.Printf("%d: Recording Started", p.frameCount)
	}
	p.recordedFrames += fmt.Sprintf("(%d:", p.frameCount)
	p.motionDetectedFrames += fmt.Sprintf("(%d:", p.frameCount-p.config.Motion.TriggerFrames+1)
}

func (p *EventLoggingRecordingListener) RecordingEnded() {
	if p.verbose {
		log.Printf("%d: Recording Ended", p.frameCount)
	}
	p.recordedFrames += fmt.Sprintf("%d)", p.frameCount)
	p.motionDetectedFrames += fmt.Sprintf("%d)", p.frameCount-p.config.Recorder.MinSecs*p.framesHz)
}

func (p *EventLoggingRecordingListener) completed() {
	if strings.HasSuffix(p.motionDetectedFrames, ":") {
		p.motionDetectedFrames += "end)"
	}

	if strings.HasSuffix(p.recordedFrames, ":") {
		p.recordedFrames += "end)"
	}

	if p.motionDetectedFrames == "" {
		p.motionDetectedFrames = "None"
	}

	if p.recordedFrames == "" {
		p.recordedFrames = "None"
	}
}

type CPTVPlaybackTester struct {
	config   *Config
	basePath string
	results  map[string]*EventLoggingRecordingListener
}

func NewCPTVPlaybackTester(conf *Config) *CPTVPlaybackTester {
	return &CPTVPlaybackTester{
		config:  conf,
		results: make(map[string]*EventLoggingRecordingListener),
	}
}

func (cpt *CPTVPlaybackTester) processIfCPTVFile(path string, info os.FileInfo, err error) error {
	if strings.HasSuffix(path, ".cptv") {
		log.Printf("Testing  %s", path)
		newResult := cpt.Detect(path)
		newResult.completed()
		shortName := path[len(cpt.basePath)+1:]
		cpt.results[shortName] = newResult
	}
	return nil
}

func (cpt *CPTVPlaybackTester) TestAllCPTVFiles(dir string) map[string]*EventLoggingRecordingListener {
	cpt.basePath = dir
	log.Printf("Looking for CPTV files in %s", cpt.basePath)
	filepath.Walk(cpt.basePath, cpt.processIfCPTVFile)
	return cpt.results
}

func (cpt *CPTVPlaybackTester) LoadAllCptvFrames(filename string) []*cptvframe.Frame {
	cpt.config.Motion.Verbose = false
	frames := make([]*cptvframe.Frame, 0, 100)

	file, reader, err := motionTesterLoadFile(filename)
	if err != nil {
		return frames[0:1]
	}
	defer file.Close()

	frame := reader.EmptyFrame()
	for {
		if err := reader.ReadFrame(frame); err != nil {
			return frames
		}
		frames = append(frames, frame)
		frame = reader.EmptyFrame()
	}
}

func (cpt *CPTVPlaybackTester) Detect(filename string) *EventLoggingRecordingListener {
	verbose := cpt.config.Motion.Verbose
	if verbose {
		log.Printf("TestFile is %s", filename)
	}

	recorder := new(recorder.NoWriteRecorder)

	file, reader, err := motionTesterLoadFile(filename)
	camera := new(TestCamera)
	listener := new(EventLoggingRecordingListener)
	listener.config = cpt.config
	listener.verbose = verbose
	listener.framesHz = camera.FPS()
	processor := motion.NewMotionProcessor(lepton3.ParseRawFrame, &cpt.config.Motion, &cpt.config.Recorder, &cpt.config.Location, listener, recorder, camera)

	if err != nil {
		log.Printf("Could not open file %v", err)
	}
	defer file.Close()

	log.Printf("Device name: %v", reader.DeviceName())
	log.Printf("Timestamp: %v", reader.Timestamp())

	fakeTime := time.Minute
	frame := reader.EmptyFrame()
	for {
		if err := reader.ReadFrame(frame); err != nil {
			if verbose {
				if err != io.EOF {
					log.Printf("Error reading file occured %v", err)
				}
				log.Printf("Last Frame gap %d", listener.frameCount-listener.lastDetection)
				log.Printf("Motion detected frames %d out of frames %d", listener.motionDetectedCount, listener.frameCount)
			}
			return listener
		}

		// The CPTV files used by the tests are missing the TimeOn
		// field so fake it avoid problems with incorrect FFC
		// detection.
		if frame.Status.TimeOn == time.Duration(0) {
			frame.Status.TimeOn = fakeTime
		}

		processor.ProcessFrame(frame)
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
