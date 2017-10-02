// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/TheCacophonyProject/thermal-recorder/output"
)

// XXX restarting camera if NextFrame dies

const framesHz = 9 // approx

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	conf, err := ConfigFromFile("thermal-recorder.toml")
	if err != nil {
		return err
	}
	log.Printf("config: %+v", *conf)

	if _, err := host.Init(); err != nil {
		return err
	}

	if err := powerupCamera(conf.PowerPin); err != nil {
		return err
	}

	camera := lepton3.New(conf.SPISpeed)
	err = camera.Open()
	if err != nil {
		return err
	}
	defer camera.Close()
	camera.SetLogFunc(func(t string) { log.Printf(t) })

	movement := new(MovementDetector)

	prevFrame := new(lepton3.Frame)
	frame := new(lepton3.Frame)

	var writer *output.FileWriter
	recordingCount := 0
	minRecordingCount := conf.MinSecs * framesHz
	for {
		err := camera.NextFrame(frame)
		if err != nil {
			return err
		}

		// If movement detected, bump the recording counter.
		if movement.Detect(frame) {
			recordingCount = minRecordingCount
		}

		// Start or stop recording if required.
		if recordingCount > 0 && writer == nil {
			filename := filepath.Join(conf.OutputDir, newRecordingTempName())
			log.Printf("recording started: %s", filename)
			writer, err = output.NewFileWriter(filename)
			if err != nil {
				return err
			}
			err = writer.WriteHeader()
			if err != nil {
				return err
			}
			// Start with an empty previous frame for a new recording.
			prevFrame = new(lepton3.Frame)
		} else if recordingCount == 0 && writer != nil {
			writer.Close()
			finalName, err := renameTempRecording(writer.Name())
			if err != nil {
				return err
			}
			log.Printf("recording stopped: %s\n", finalName)
			writer = nil
		}

		// If recording, write the frame.
		if writer != nil {
			err := writer.WriteFrame(prevFrame, frame)
			if err != nil {
				return err
			}
			recordingCount--
		}

		frame, prevFrame = prevFrame, frame
	}

	return nil
}

func newRecordingTempName() string {
	return time.Now().Format("20060102.150405.000.cptv.temp")
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

func powerupCamera(pin string) error {
	// XXX can we detect if the camera was already powered up? If it was off sleep for a few seconds.
	powerPin := gpioreg.ByName(pin)
	if powerPin == nil {
		return errors.New("unable to load power pin")
	}
	if err := powerPin.Out(gpio.High); err != nil {
		return fmt.Errorf("failed to set camera power pin: %v", err)
	}
	return nil
}
