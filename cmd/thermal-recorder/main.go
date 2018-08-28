// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"math"
	"net"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sync/atomic"
	"syscall"
	"time"

	cptv "github.com/TheCacophonyProject/go-cptv"
	"github.com/TheCacophonyProject/lepton3"
	arg "github.com/alexflint/go-arg"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

const (
	framesHz    = 9 // approx
	cptvTempExt = "cptv.temp"

	frameLogIntervalFirstMin = 15 * framesHz
	frameLogInterval         = 60 * 5 * framesHz
)

var (
	version         = "<not set>"
	snapshotChan    = make(chan error, 1)
	snapshotPending atomic.Value
)

type Args struct {
	ConfigFile         string `arg:"-c,--config" help:"path to configuration file"`
	UploaderConfigFile string `arg:"-u,--uploader-config" help:"path to uploader config file"`
	Timestamps         bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
}

func (Args) Version() string {
	return version
}

func procArgs() Args {
	var args Args
	args.ConfigFile = "/etc/thermal-recorder.yaml"
	args.UploaderConfigFile = "/etc/thermal-uploader.yaml"
	arg.MustParse(&args)
	return args
}

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	args := procArgs()
	if !args.Timestamps {
		log.SetFlags(0) // Removes default timestamp flag
	}

	log.Printf("running version: %s", version)
	conf, err := ParseConfigFiles(args.ConfigFile, args.UploaderConfigFile)
	if err != nil {
		return err
	}
	logConfig(conf)

	log.Println("starting d-bus service")
	snapshotPending.Store(false)
	err = startService()
	if err != nil {
		return err
	}

	log.Println("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

	turret := NewTurretController(conf.Turret)
	go turret.Start()

	log.Println("deleting temp files")
	if err := deleteTempFiles(conf.OutputDir); err != nil {
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

	recordingLed := gpioreg.ByName(conf.LEDs.Recording)
	if recordingLed == nil {
		return fmt.Errorf("failed to load pin: %s", conf.LEDs.Recording)
	}
	if err := recordingLed.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set recording LED off: %v", err)
	}

	for {
		// Set up listener for frames sent by leptond.
		os.Remove(conf.FrameInput)
		listener, err := net.Listen("unixpacket", conf.FrameInput)
		if err != nil {
			return err
		}
		log.Print("waiting for camera connection")

		conn, err := listener.Accept()
		if err != nil {
			log.Printf("socket accept failed: %v", err)
			continue
		}

		// Prevent concurrent connections.
		listener.Close()

		err = handleConn(conn, conf, turret, recordingLed)
		log.Printf("camera connection ended with: %v", err)
	}
}

func handleConn(conn net.Conn, conf *Config, turret *TurretController, recordingLed gpio.PinIO) error {
	defer recordingLed.Out(gpio.Low)

	totalFrames := 0

	motionLogFrame := -999

	minFrames := conf.MinSecs * framesHz
	maxFrames := conf.MaxSecs * framesHz
	framesWritten := 0
	lastFrame := 0

	motion := NewMotionDetector(conf.Motion)
	window := NewWindow(conf.WindowStart, conf.WindowEnd)

	loopFrames := conf.PreviewSecs * framesHz
	frameLoop := NewFrameLoop(loopFrames)

	frame := frameLoop.Current()

	var writer *cptv.FileWriter
	defer func() {
		if writer != nil {
			writer.Close()
			os.Remove(writer.Name())
		}
	}()

	rawFrame := new(lepton3.RawFrame)

	log.Print("new camera connection, reading frames")

	for {
		_, err := conn.Read(rawFrame[:])
		if err != nil {
			return err
		}
		rawFrame.ToFrame(frame)
		totalFrames++

		if snapshotPending.Load() == true {
			snapshotChan <- frameToSnapshot(conf.OutputDir, frame)
		}

		if totalFrames%frameLogIntervalFirstMin == 0 &&
			totalFrames <= 60*framesHz || totalFrames%frameLogInterval == 0 {
			log.Printf("%d frames for this connection", totalFrames)
		}

		// If motion detected, allow minFrames more frames.
		if motion.Detect(frame) {
			turret.Update(motion)
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
				lastFrame = min(framesWritten+minFrames, maxFrames)
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
			err = writer.WriteHeader(conf.DeviceName)
			if err != nil {
				return err
			}

			if err = writeInitialFramesToFile(writer, frameLoop.GetHistory()); err != nil {
				return err
			}

		} else if writer != nil && framesWritten > lastFrame {
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
			framesWritten = 0
			lastFrame = 0
		}

		// If recording, write the frame.
		if writer != nil {
			err := writer.WriteFrame(frame)
			if err != nil {
				return err
			}
			framesWritten++
		}

		frame = frameLoop.Move()
	}
}

func logConfig(conf *Config) {
	log.Printf("device name: %s", conf.DeviceName)
	log.Printf("frame input: %s", conf.FrameInput)
	log.Printf("output dir: %s", conf.OutputDir)
	log.Printf("recording limits: %ds to %ds", conf.MinSecs, conf.MaxSecs)
	log.Printf("preview seconds: %d", conf.PreviewSecs)
	log.Printf("minimum disk space: %d", conf.MinDiskSpace)
	log.Printf("motion: %+v", conf.Motion)
	log.Printf("leds: %+v", conf.LEDs)
	if !conf.WindowStart.IsZero() {
		log.Printf("recording window: %02d:%02d to %02d:%02d",
			conf.WindowStart.Hour(), conf.WindowStart.Minute(),
			conf.WindowEnd.Hour(), conf.WindowEnd.Minute())
	}
	if conf.Turret.Active {
		log.Printf("Turret active")
		log.Printf("\tPID: %v", conf.Turret.PID)
		log.Printf("\tServoX: %+v", conf.Turret.ServoX)
		log.Printf("\tServoY: %+v", conf.Turret.ServoY)
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

func writeInitialFramesToFile(writer *cptv.FileWriter, frames []*lepton3.Frame) error {
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

func newSnapshot() error {
	if snapshotPending.Load() == true {
		//TODO return the pending snapshot error also.
		return errors.New("snapshot already pending")
	}
	log.Println("take new snapshot")
	snapshotPending.Store(true)
	defer snapshotPending.Store(false)
	var err error
	select {
	case err = <-snapshotChan:
	case <-time.After(10 * time.Second):
		err = errors.New("timeout for snapshot")
	}
	return err
}

func frameToSnapshot(dir string, f *lepton3.Frame) error {
	g16 := image.NewGray16(image.Rect(0, 0, 160, 120))
	// Max and min are needed for normalization of the frame
	var valMax uint16
	var valMin uint16 = 65535
	for _, row := range f {
		for _, val := range row {
			valMax = uint16(math.Max(float64(valMax), float64(val)))
			valMin = uint16(math.Min(float64(valMin), float64(val)))
		}
	}

	var norm = 65535 / (valMax - valMin)
	for y, row := range f {
		for x, val := range row {
			g16.SetGray16(x, y, color.Gray16{Y: (val - valMin) * norm})
		}
	}

	out, err := os.Create(path.Join(dir, "still.png"))
	if err != nil {
		return err
	}
	defer out.Close()
	return png.Encode(out, g16)
}
