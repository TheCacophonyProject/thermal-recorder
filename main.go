// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"syscall"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	arg "github.com/alexflint/go-arg"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	cptv "github.com/TheCacophonyProject/go-cptv"
)

const framesHz = 9 // approx
const cptvTempExt = "cptv.temp"

type Args struct {
	ConfigFile string `arg:"-c,--config" help:"path to configuration file"`
	Quick      bool   `arg:"-q,--quick" help:"don't cycle camera power on startup"`
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
	log.SetFlags(0) // Removes default timestamp flag

	args := procArgs()
	conf, err := ParseConfigFile(args.ConfigFile)
	if err != nil {
		return err
	}
	logConfig(conf)

	log.Println("deleting temp files")
	if err := deleteTempFiles(conf.OutputDir); err != nil {
		return err
	}

	log.Println("checking if there is enough free disk space for recording")
	if ok, err := checkDiskSpace(conf.MinDiskSpace, conf.OutputDir); err != nil {
		return err
	} else if !ok {
		log.Printf("not enough free disk space for recording")
	}

	log.Println("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

	runningLed := gpioreg.ByName(conf.LEDs.Running)
	if runningLed == nil {
		return fmt.Errorf("failed to load pin: %s", conf.LEDs.Running)
	}
	if err := runningLed.Out(gpio.High); err != nil {
		return fmt.Errorf("failed to set running led on: %v", err)
	}
	defer runningLed.Out(gpio.Low)

	if !args.Quick {
		if err := cycleCameraPower(conf.PowerPin); err != nil {
			return err
		}
	}

	var camera *lepton3.Lepton3
	defer func() {
		if camera != nil {
			camera.Close()
		}
	}()
	for {
		camera = lepton3.New(conf.SPISpeed)
		camera.SetLogFunc(func(t string) { log.Printf(t) })

		log.Println("opening camera")
		if err := camera.Open(); err != nil {
			return err
		}

		log.Println("enabling radiometry")
		if err := camera.SetRadiometry(true); err != nil {
			return err
		}

		err := runRecordings(conf, camera)
		if err != nil {
			if _, isNextFrameErr := err.(*nextFrameErr); !isNextFrameErr {
				return err
			}
			log.Printf("recording error: %v", err)
		}

		log.Println("closing camera")
		camera.Close()

		err = cycleCameraPower(conf.PowerPin)
		if err != nil {
			return err
		}
	}

	return nil
}

func runRecordings(conf *Config, camera *lepton3.Lepton3) error {
	motion := NewMotionDetector(conf.Motion)

	prevFrame := new(lepton3.Frame)
	frame := new(lepton3.Frame)

	var writer *cptv.FileWriter
	defer func() {
		if writer != nil {
			writer.Close()
			os.Remove(writer.Name())
		}
	}()

	window := NewWindow(conf.WindowStart, conf.WindowEnd)

	log.Println("reading frames")

	totalFrames := 0
	const frameLogIntervalFirstMin = 15 * framesHz
	const frameLogInterval = 60 * 5 * framesHz

	motionLogFrame := -999

	minFrames := conf.MinSecs * framesHz
	maxFrames := conf.MaxSecs * framesHz
	numFrames := 0
	lastFrame := 0

	recordingLed := gpioreg.ByName(conf.LEDs.Recording)
	if recordingLed == nil {
		return fmt.Errorf("failed to load pin: %s", conf.LEDs.Recording)
	}
	if err := recordingLed.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set recording LED off: %v", err)
	}
	defer recordingLed.Out(gpio.Low)
	for {
		err := camera.NextFrame(frame)
		if err != nil {
			return &nextFrameErr{err}
		}
		totalFrames++
		if totalFrames%frameLogIntervalFirstMin == 0 &&
			totalFrames <= 60*framesHz || totalFrames%frameLogInterval == 0 {
			log.Printf("%d frames seen", totalFrames)
		}

		// If motion detected, allow minFrames more frames.
		if motion.Detect(frame) {
			shouldLogMotion := motionLogFrame <= totalFrames-(10*framesHz)
			if shouldLogMotion {
				motionLogFrame = totalFrames
			}
			if !window.Active() {
				if shouldLogMotion {
					log.Print("motion detected but outside of recording window")
				}
			} else if enoughSpace, err := checkDiskSpace(conf.MinDiskSpace, conf.OutputDir); err != nil {
				return err
			} else if !enoughSpace {
				if shouldLogMotion {
					log.Print("motion detected but not enough free disk space to start recording")
				}
			} else {
				lastFrame = min(numFrames+minFrames, maxFrames)
			}
		}

		// Start or stop recording if required.
		if lastFrame > 0 && writer == nil {
			filename := filepath.Join(conf.OutputDir, newRecordingTempName())
			log.Printf("recording started: %s", filename)
			if err := recordingLed.Out(gpio.High); err != nil {
				return fmt.Errorf("failed to set recording LED on: %v", err)
			}
			writer, err = cptv.NewFileWriter(filename)
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
			if err := recordingLed.Out(gpio.Low); err != nil {
				return fmt.Errorf("failed to set recording LED off: %v", err)
			}
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

func logConfig(conf *Config) {
	log.Printf("SPI speed: %d", conf.SPISpeed)
	log.Printf("power pin: %s", conf.PowerPin)
	log.Printf("output dir: %s", conf.OutputDir)
	log.Printf("recording limits: %ds to %ds", conf.MinSecs, conf.MaxSecs)
	log.Printf("minimum disk space: %d", conf.MinDiskSpace)
	log.Printf("motion: %+v", conf.Motion)
	log.Printf("leds: %+v", conf.LEDs)
	if !conf.WindowStart.IsZero() {
		log.Printf("recording window: %02d:%02d to %02d:%02d",
			conf.WindowStart.Hour(), conf.WindowStart.Minute(),
			conf.WindowEnd.Hour(), conf.WindowEnd.Minute())
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

func cycleCameraPower(pinName string) error {
	if pinName == "" {
		return nil
	}

	pin := gpioreg.ByName(pinName)

	log.Println("turning camera power off")
	if err := pin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set camera power pin low: %v", err)
	}
	time.Sleep(2 * time.Second)

	log.Println("turning camera power on")
	if err := pin.Out(gpio.High); err != nil {
		return fmt.Errorf("failed to set camera power pin high: %v", err)
	}

	log.Println("waiting for camera startup")
	time.Sleep(8 * time.Second)
	log.Println("camera should be ready")
	return nil
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
