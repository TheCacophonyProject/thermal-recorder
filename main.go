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
	arg "github.com/alexflint/go-arg"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/TheCacophonyProject/thermal-recorder/output"
)

const framesHz = 9 // approx

type Args struct {
	ConfigFile string `arg:"-c,--config" help:"path to configuration file"`
}

func procArgs() Args {
	var args Args
	args.ConfigFile = "/etc/thermal-recorder.yaml"
	arg.MustParse(&args)
	return args
}

type nextFrameErr struct {
	cause error
}

func (e *nextFrameErr) Error() string {
	return e.cause.Error()
}

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	args := procArgs()
	conf, err := ParseConfigFile(args.ConfigFile)
	if err != nil {
		return err
	}
	log.Printf("config: %+v", *conf)

	log.Println("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

	log.Println("powering up camera")
	if err := powerupCamera(conf.PowerPin); err != nil {
		return err
	}

	var camera *lepton3.Lepton3
	defer camera.Close()

	for {
		log.Println("opening camera")
		camera = lepton3.New(conf.SPISpeed)
		if err := camera.Open(); err != nil {
			return err
		}
		log.Println("enabling radiometry")
		if err := camera.SetRadiometry(true); err != nil {
			return err
		}
		camera.SetLogFunc(func(t string) { log.Printf(t) })

		err := runRecordings(conf, camera)
		if err != nil {
			if _, isNextFrameErr := err.(*nextFrameErr); !isNextFrameErr {
				return err
			}
		}
		log.Println("closing camera")
		camera.Close()

		log.Println("cycling camera power")
		err = cameraPowerOffOn(conf.PowerPin)
		if err != nil {
			return err
		}
	}

	return nil
}

func runRecordings(conf *Config, camera *lepton3.Lepton3) error {
	motion := NewMotionDetector(
		conf.Motion.DeltaThresh,
		conf.Motion.CountThresh,
		conf.Motion.TempThresh,
	)

	prevFrame := new(lepton3.Frame)
	frame := new(lepton3.Frame)

	var writer *output.FileWriter
	defer func() {
		if writer != nil {
			writer.Close()
			os.Remove(writer.Name())
		}
	}()

	log.Println("reading frames")

	totalFrames := 0
	const frameLogInterval = 15 * framesHz

	minFrames := conf.MinSecs * framesHz
	maxFrames := conf.MaxSecs * framesHz
	numFrames := 0
	lastFrame := 0
	for {
		err := camera.NextFrame(frame)
		if err != nil {
			return &nextFrameErr{err}
		}
		totalFrames++
		if totalFrames%frameLogInterval == 0 {
			log.Printf("%d frames seen", totalFrames)
		}

		// If motion detected, allow minFrames more frames.
		if motion.Detect(frame) {
			lastFrame = min(numFrames+minFrames, maxFrames)
		}

		// Start or stop recording if required.
		if lastFrame > 0 && writer == nil {
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
		} else if writer != nil && numFrames > lastFrame {
			writer.Close()
			finalName, err := renameTempRecording(writer.Name())
			if err != nil {
				return err
			}
			log.Printf("recording stopped: %s\n", finalName)
			writer = nil
			numFrames = 0
			lastFrame = 0
		}

		// If recording, write the frame.
		if writer != nil {
			err := writer.WriteFrame(prevFrame, frame)
			if err != nil {
				return err
			}
			numFrames++
		}

		frame, prevFrame = prevFrame, frame
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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

func cameraPowerOffOn(pin string) error {
	powerPin := gpioreg.ByName(pin)
	if err := powerPin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set camera power pin low: %v", err)
	}
	time.Sleep(2 * time.Second)
	if err := powerPin.Out(gpio.High); err != nil {
		return fmt.Errorf("failed to set camera power pin high: %v", err)
	}
	time.Sleep(6 * time.Second)
	return nil
}

func powerupCamera(pin string) error {
	powerPin := gpioreg.ByName(pin)
	if powerPin == nil {
		return errors.New("unable to load power pin")
	}
	if err := powerPin.Out(gpio.High); err != nil {
		return fmt.Errorf("failed to set camera power pin: %v", err)
	}
	return nil
}
