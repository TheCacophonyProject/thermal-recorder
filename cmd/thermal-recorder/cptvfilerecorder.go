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
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	cptv "github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	yaml "gopkg.in/yaml.v2"
)

func NewCPTVFileRecorder(config *Config, camera cptvframe.CameraSpec, brand, model string) *CPTVFileRecorder {
	motionYAML, err := yaml.Marshal(config.Motion)
	if err != nil {
		panic(fmt.Sprintf("failed to convert motion config to YAML: %v", err))
	}
	cptvHeader := cptv.Header{
		DeviceName:   config.DeviceName,
		PreviewSecs:  config.Recorder.PreviewSecs,
		MotionConfig: string(motionYAML),
		Latitude:     config.Location.Latitude,
		Longitude:    config.Location.Longitude,
		LocTimestamp: config.Location.Timestamp,
		Altitude:     config.Location.Altitude,
		Accuracy:     config.Location.Accuracy,
		FPS:          camera.FPS(),
		Brand:        brand,
		Model:        model,
	}
	if config.DeviceID > 0 {
		cptvHeader.DeviceID = config.DeviceID
	}
	return &CPTVFileRecorder{
		outputDir:    config.OutputDir,
		header:       cptvHeader,
		minDiskSpace: config.MinDiskSpace,
		camera:       camera,
	}
}

type CPTVFileRecorder struct {
	outputDir    string
	header       cptv.Header
	minDiskSpace uint64
	camera       cptvframe.CameraSpec
	writer       *cptv.FileWriter
}

func (cfr *CPTVFileRecorder) CheckCanRecord() error {
	enoughSpace, err := checkDiskSpace(cfr.minDiskSpace, cfr.outputDir)
	if err != nil {
		return fmt.Errorf("Problem with checking disk space: %v", err)
	} else if !enoughSpace {
		return errors.New("Motion detected but not enough free disk space to start recording")
	}
	return nil
}

func (fw *CPTVFileRecorder) StartRecording() error {
	filename := filepath.Join(fw.outputDir, newRecordingTempName())
	log.Printf("recording started: %s", filename)

	writer, err := cptv.NewFileWriter(filename, fw.camera)
	if err != nil {
		return err
	}

	if err = writer.WriteHeader(fw.header); err != nil {
		writer.Close()
		return err
	}

	fw.writer = writer
	return nil
}

func (fw *CPTVFileRecorder) StopRecording() error {
	if fw.writer != nil {
		fw.writer.Close()

		finalName, err := renameTempRecording(fw.writer.Name())
		log.Printf("recording stopped: %s\n", finalName)
		fw.writer = nil

		return err
	}
	return nil
}

func (fw *CPTVFileRecorder) Stop() {
	if fw.writer != nil {
		fw.writer.Close()
		os.Remove(fw.writer.Name())
		fw.writer = nil
	}
}

func (fw *CPTVFileRecorder) WriteFrame(frame *cptvframe.Frame) error {
	return fw.writer.WriteFrame(frame)
}

func newRecordingTempName() string {
	return time.Now().Format("20060102.150405.000." + cptvTempExt)
}

func renameTempRecording(tempName string) (string, error) {
	finalName := recordingFinalName(tempName)
	err := os.Rename(tempName, finalName)
	if err != nil {
		return "", err
	}
	return finalName, nil
}

var reTempName = regexp.MustCompile(`(.+)\.temp$`)

func recordingFinalName(filename string) string {
	return reTempName.ReplaceAllString(filename, `$1`)
}

func deleteTempFiles(directory string) error {
	matches, _ := filepath.Glob(filepath.Join(directory, "*."+cptvTempExt))
	for _, filename := range matches {
		if err := os.Remove(filename); err != nil {
			return err
		}
	}
	return nil
}

func checkDiskSpace(mb uint64, dir string) (bool, error) {
	var fs syscall.Statfs_t
	if err := syscall.Statfs(dir, &fs); err != nil {
		return false, err
	}
	return fs.Bavail*uint64(fs.Bsize)/1024/1024 >= mb, nil
}
