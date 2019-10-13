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
	"log"
	"net"
	"os"

	arg "github.com/alexflint/go-arg"
	"periph.io/x/periph/host"

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/TheCacophonyProject/thermal-recorder/throttle"
)

const (
	framesHz    = lepton3.FramesHz // approx
	cptvTempExt = "cptv.temp"

	frameLogIntervalFirstMin = 15 * framesHz
	frameLogInterval         = 60 * 5 * framesHz
)

var (
	version   = "<not set>"
	processor *motion.MotionProcessor
)

type Args struct {
	ConfigDir    string `arg:"-c,--config" help:"path to configuration directory"`
	Timestamps   bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
	TestCptvFile string `arg:"-f, --testfile" help:"Run a CPTV file through to see what the results are"`
	Verbose      bool   `arg:"-v, --verbose" help:"Make logging more verbose"`
}

func (Args) Version() string {
	return version
}

func procArgs() Args {
	var args Args
	args.ConfigDir = config.DefaultConfigDir
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
	conf, err := ParseConfig(args.ConfigDir)
	if err != nil {
		return err
	}

	logConfig(conf)

	if args.TestCptvFile != "" {
		conf.Motion.Verbose = args.Verbose
		results := NewCPTVPlaybackTester(conf).Detect(args.TestCptvFile)
		log.Printf("Detected: %-16s Recorded: %-16s Motion frames: %d/%d", results.motionDetectedFrames, results.recordedFrames, results.motionDetectedCount, results.frameCount)
		return nil
	}

	log.Println("starting d-bus service")
	err = startService(conf.OutputDir)
	if err != nil {
		return err
	}

	deleteSnapshot(conf.OutputDir)

	log.Println("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

	log.Println("deleting temp files")
	if err := deleteTempFiles(conf.OutputDir); err != nil {
		return err
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

		err = handleConn(conn, conf)
		log.Printf("camera connection ended with: %v", err)
	}
}

func handleConn(conn net.Conn, conf *Config) error {

	totalFrames := 0

	cptvRecorder := NewCPTVFileRecorder(conf)
	defer cptvRecorder.Stop()
	var recorder recorder.Recorder = cptvRecorder

	if conf.Throttler.Activate {
		minRecordingLength := conf.Recorder.MinSecs + conf.Recorder.PreviewSecs
		recorder = throttle.NewThrottledRecorder(cptvRecorder, &conf.Throttler, minRecordingLength, new(throttle.ThrottledEventRecorder))
	}

	processor = motion.NewMotionProcessor(&conf.Motion, &conf.Recorder, &conf.Location, nil, recorder)

	rawFrame := new(lepton3.RawFrame)

	log.Print("new camera connection, reading frames")

	for {
		_, err := conn.Read(rawFrame[:])
		if err != nil {
			return err
		}
		totalFrames++

		if totalFrames%frameLogIntervalFirstMin == 0 &&
			totalFrames <= 60*framesHz || totalFrames%frameLogInterval == 0 {
			log.Printf("%d frames for this connection", totalFrames)
		}

		processor.Process(rawFrame)
	}
}

func logConfig(conf *Config) {
	log.Printf("device name: %s", conf.DeviceName)
	log.Printf("frame input: %s", conf.FrameInput)
	log.Printf("output dir: %s", conf.OutputDir)
	log.Printf("recording limits: %ds to %ds", conf.Recorder.MinSecs, conf.Recorder.MaxSecs)
	log.Printf("preview seconds: %d", conf.Recorder.PreviewSecs)
	log.Printf("minimum disk space: %d", conf.MinDiskSpace)
	log.Printf("motion: %+v", conf.Motion)
	log.Printf("throttler: %+v", conf.Throttler)
	log.Printf("location latitude: %v", conf.Location.Latitude)
	log.Printf("location longitude: %v", conf.Location.Longitude)
	log.Printf("recording window: %s", conf.Recorder.Window)
}
