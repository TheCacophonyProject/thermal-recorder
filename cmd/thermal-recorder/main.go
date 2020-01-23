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
	"bytes"
	"io"
	"log"
	"net"
	"os"
	"strings"

	arg "github.com/alexflint/go-arg"
	"gopkg.in/yaml.v1"
	"periph.io/x/periph/host"

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/headers"
	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/TheCacophonyProject/thermal-recorder/throttle"
)

const (
	cptvTempExt = "cptv.temp"
)

var (
	version                  = "<not set>"
	processor                *motion.MotionProcessor
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

type HeaderInfo struct {
	resX      int
	resY      int
	fps       int
	framesize int
	model     string
	brand     string
}

func (h *HeaderInfo) ResX() int {
	return h.resX
}
func (h *HeaderInfo) ResY() int {
	return h.resY
}
func (h *HeaderInfo) FPS() int {
	return h.fps
}

func NewHeader(headerInfo map[string]interface{}) *HeaderInfo {
	return &HeaderInfo{
		resX:      headerInfo[headers.XResolution].(int),
		resY:      headerInfo[headers.YResolution].(int),
		fps:       headerInfo[headers.FPS].(int),
		framesize: headerInfo[headers.FrameSize].(int),
		model:     headerInfo[headers.Model].(string),
		brand:     headerInfo[headers.Brand].(string),
	}
}

func readHeader(reader *bufio.Reader) (*HeaderInfo, error) {
	var buf bytes.Buffer
	for {
		line, err := reader.ReadString(byte('\n'))
		if err != nil {
			return nil, err
		}
		if strings.Trim(line, " ") == "\n" {
			break
		}
		buf.WriteString(line)
	}
	header := make(map[string]interface{})
	err := yaml.Unmarshal(buf.Bytes(), &header)
	return NewHeader(header), err
}

func handleConn(conn net.Conn, conf *Config) error {

	totalFrames := 0
	reader := bufio.NewReader(conn)
	header, err := readHeader(reader)
	if err != nil {
		return err
	}
	cptvRecorder := NewCPTVFileRecorder(conf, header, header.brand, header.model)
	defer cptvRecorder.Stop()
	var recorder recorder.Recorder = cptvRecorder

	if conf.Throttler.Activate {
		minRecordingLength := conf.Recorder.MinSecs + conf.Recorder.PreviewSecs
		recorder = throttle.NewThrottledRecorder(cptvRecorder, &conf.Throttler, minRecordingLength, new(throttle.ThrottledEventRecorder), header)
	}

	processor = motion.NewMotionProcessor(&conf.Motion, &conf.Recorder, &conf.Location, nil, recorder, header)

	log.Print("new camera connection, reading frames")

	frameLogIntervalFirstMin *= header.FPS()
	frameLogInterval *= header.FPS()

	rawFrame := new(lepton3.RawFrame)
	for {
		_, err := io.ReadFull(reader, rawFrame[:])
		if err != nil {
			return err
		}
		totalFrames++

		if totalFrames%frameLogIntervalFirstMin == 0 &&
			totalFrames <= 60*header.FPS() || totalFrames%frameLogInterval == 0 {
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
