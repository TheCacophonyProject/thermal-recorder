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
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/TheCacophonyProject/event-reporter/eventclient"
	goconfig "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/window"
	arg "github.com/alexflint/go-arg"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/rjeczalik/notify"
	"periph.io/x/periph/host"

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/thermal-recorder/headers"
	"github.com/TheCacophonyProject/thermal-recorder/leptondController"
	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/TheCacophonyProject/thermal-recorder/throttle"
)

const (
	cptvTempExt = "cptv.temp"
	clearBuffer = "clear"
)

var (
	version    = "<not set>"
	processor  *motion.MotionProcessor
	headerInfo *headers.HeaderInfo = nil

	frameLogIntervalFirstMin = 15
	frameLogInterval         = 60 * 5
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

// checkConfigChanges will compare the config from when first loaded to a new config each time
// the config file is modified.
// If there is a difference then the program will exit and systemd will restart the service, causing
// the new config to be loaded.
func checkConfigChanges(conf *Config, configDir string) error {
	configFilePath := filepath.Join(configDir, config.ConfigFileName)
	fsEvents := make(chan notify.EventInfo, 1)
	if err := notify.Watch(configFilePath, fsEvents, notify.InCloseWrite, notify.InMovedTo); err != nil {
		return err
	}
	defer notify.Stop(fsEvents)

	for {
		<-fsEvents
		newConfig, err := ParseConfig(configDir)
		if err != nil {
			log.Println("error reloading config:", err)
			continue
		}

		// Need to set Window.Now in Recorder.Config to nil to compare them.
		isRecorderConfigEqual := func(x, y recorder.RecorderConfig) bool {
			x.Window.Now = nil
			y.Window.Now = nil
			return cmp.Equal(x, y, cmp.AllowUnexported(window.Window{}))
		}

		// Checking the diff of the current and new config.
		diff := cmp.Diff(
			conf,
			newConfig,
			cmp.AllowUnexported(
				config.Location{},
				recorder.RecorderConfig{},
				goconfig.ThermalMotion{},
				goconfig.ThermalThrottler{},
				window.Window{},
			),
			cmpopts.IgnoreFields(Config{}, "Motion"), // Ignore the motion config as this is modified with config.LoadMotionConfig
			cmp.Comparer(isRecorderConfigEqual))      // Custom compare function for recorder config ignoring Window.Now

		if diff != "" {
			log.Println("Config changed:", diff, "\nExiting to allow systemctl to restart service.")
			os.Exit(0)
		} else {
			log.Println("No relevant changes detected in config file.")
		}

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

	// Check for config changes.
	go checkConfigChanges(conf, args.ConfigDir)

	conf.Verbose = args.Verbose

	if args.TestCptvFile != "" {
		results := NewCPTVPlaybackTester(conf).Detect(args.TestCptvFile)
		logConfig(conf)

		log.Printf("Detected: %-16s Recorded: %-16s Motion frames: %d/%d", results.motionDetectedFrames, results.recordedFrames, results.motionDetectedCount, results.frameCount)
		return nil
	}

	if _, err := os.Stat(conf.OutputDir); os.IsNotExist(err) {
		return os.MkdirAll(conf.OutputDir, 0755)
	}

	log.Println("starting d-bus service")
	err = startService(conf.OutputDir)
	if err != nil {
		return err
	}

	log.Println("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

	log.Println("deleting temp files")
	if err := deleteTempFiles(conf.OutputDir); err != nil {
		return err
	}

	go snapshotRecordingTriggers(conf.Recorder.Window)

	for {
		// Set up listener for frames sent by leptond.
		os.Remove(conf.FrameInput)
		listener, err := net.Listen("unix", conf.FrameInput)
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
	leptondController.SetAutoFFC(true)
	totalFrames := 0
	reader := bufio.NewReader(conn)
	var err error
	headerInfo, err = headers.ReadHeaderInfo(reader)
	if err != nil {
		return err
	}

	log.Printf("connection from %s %s (%dx%d@%dfps)", headerInfo.Brand(), headerInfo.Model(), headerInfo.ResX(), headerInfo.ResY(), headerInfo.FPS())
	conf.LoadMotionConfig(headerInfo.Model())
	logConfig(conf)

	parseFrame := frameParser(headerInfo.Brand(), headerInfo.Model())
	if parseFrame == nil {
		return fmt.Errorf("unable to handle frames for %s %s", headerInfo.Brand(), headerInfo.Model())
	}

	cptvRecorder := NewCPTVFileRecorder(conf, headerInfo, headerInfo.Brand(), headerInfo.Model(), headerInfo.CameraSerial(), headerInfo.Firmware())
	defer cptvRecorder.Stop()
	var recorder recorder.Recorder = cptvRecorder

	if conf.Throttler.Activate {
		minRecordingLength := conf.Recorder.MinSecs + conf.Recorder.PreviewSecs
		recorder = throttle.NewThrottledRecorder(cptvRecorder, &conf.Throttler, minRecordingLength, new(throttle.ThrottledEventRecorder), headerInfo)
	}

	// Constant Recorder
	var constantRecorder *CPTVFileRecorder
	if conf.Recorder.ConstantRecorder {
		constantRecorder = NewCPTVFileRecorder(conf, headerInfo, headerInfo.Brand(), headerInfo.Model(), headerInfo.CameraSerial(), headerInfo.Firmware())
		constantRecorder.SetAsConstantRecorder()
	}

	processor = motion.NewMotionProcessor(
		parseFrame,
		&conf.Motion,
		&conf.Recorder,
		&conf.Location,
		nil,
		recorder,
		headerInfo,
		constantRecorder,
		NewCPTVFileRecorder(conf, headerInfo, headerInfo.Brand(), headerInfo.Model(), headerInfo.CameraSerial(), headerInfo.Firmware()),
	)

	log.Print("reading frames")

	frameLogIntervalFirstMin *= headerInfo.FPS()
	frameLogInterval *= headerInfo.FPS()
	rawFrame := make([]byte, headerInfo.FrameSize())
	for {
		_, err := io.ReadFull(reader, rawFrame[:5])
		if err != nil {
			return err
		}
		message := string(rawFrame[:5])
		if message == clearBuffer {
			log.Print("clearing motion buffer")
			processor.Reset(headerInfo)
			continue
		}

		_, err = io.ReadFull(reader, rawFrame[5:])
		if err != nil {
			return err
		}
		message = string(rawFrame[:5])
		totalFrames++

		if totalFrames%frameLogIntervalFirstMin == 0 &&
			totalFrames <= 60*headerInfo.FPS() || totalFrames%frameLogInterval == 0 {
			log.Printf("%d frames for this connection", totalFrames)
		}

		err = processor.Process(rawFrame)
		if _, isBadFrame := err.(*lepton3.BadFrameErr); isBadFrame {
			event := eventclient.Event{
				Timestamp: time.Now(),
				Type:      "bad-thermal-frame",
				Details:   map[string]interface{}{"description": map[string]interface{}{"details": err.Error()}},
			}
			eventclient.AddEvent(event)
			log.Println("bad frame deteccted, requesting camera to restart")
			leptondController.RestartCamera()
		}
	}
}

func frameParser(brand, model string) func([]byte, *cptvframe.Frame, int) error {
	if brand != "flir" {
		return nil
	}
	switch model {
	case lepton3.Model, lepton3.Model35:
		return lepton3.ParseRawFrame
	case "boson":
		return convertRawBosonFrame
	}
	return nil
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
