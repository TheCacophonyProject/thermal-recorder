// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/TheCacophonyProject/thermal_recorder/output"
)

// XXX restarting camera if NextFrame dies

// XXX these should come from a configuration file
const (
	spiSpeed         = 30000000
	powerPin         = "GPIO23"
	directory        = "/mnt/ramdisk"
	minRecordingSecs = 10
	cameraHz         = 9 // approx
)

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	_, err := host.Init()
	if err != nil {
		return err
	}

	if err := powerupCamera(powerPin); err != nil {
		return err
	}

	camera := lepton3.New(spiSpeed)
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
	const minRecordingCount = minRecordingSecs * cameraHz
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
			log.Println("recording started")
			writer, err = output.NewFileWriter(newRecordingName())
			if err != nil {
				return err
			}
			err = writer.WriteHeader()
			if err != nil {
				return err
			}
		} else if recordingCount == 0 && writer != nil {
			log.Println("recording stopped")
			writer.Close()
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

func newRecordingName() string {
	basename := time.Now().Format("20060102.150405.000.cptv")
	return filepath.Join(directory, basename)
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
