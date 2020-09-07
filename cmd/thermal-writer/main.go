// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2020, The Cacophony Project
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
	//"math"
	"bufio"
	"encoding/binary"
	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/thermal-recorder/headers"
	arg "github.com/alexflint/go-arg"
	"github.com/gonutz/framebuffer"
	"image"
	"image/draw"
	"io"
	"log"
	"net"
	"os"
	"time"
)

var (
	version                  = "<not set>"
	frameLogIntervalFirstMin = 15
	frameLogInterval         = 60 * 5
)

const newFileInterval = time.Minute

type Args struct {
	ConfigDir  string `arg:"-c,--config" help:"path to configuration directory"`
	Timestamps bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
	FrameRate  bool   `arg:"-r,--frame-rate" help:"log frame rate"`
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

	// Setup HDMI output:

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

		listener.Close() // Prevent concurrent connections.

		err = handleConn(conn, conf, args.FrameRate)
		log.Printf("camera connection ended with: %v", err)
	}
}

func handleConn(conn net.Conn, conf *Config, logFrameRate bool) error {
	gray := image.NewGray(image.Rect(0, 0, 640, 512))
	fb, err := framebuffer.Open("/dev/fb0")
	if err != nil {
		panic(err)
	}
	defer fb.Close()

	totalFrames := 0
	reader := bufio.NewReader(conn)
	header, err := headers.ReadHeaderInfo(reader)
	if err != nil {
		return err
	}

	log.Printf("connection from %s %s (%dx%d@%dfps)", header.Brand(), header.Model(), header.ResX(), header.ResY(), header.FPS())

	const inFlight = 256

	writeFrames := make(chan []byte, inFlight)
	spentFrames := make(chan []byte, inFlight)
	for i := 0; i < inFlight; i++ {
		spentFrames <- make([]byte, header.FrameSize())
	}

	go writer(writeFrames, conf, header, spentFrames)

	log.Print("reading frames")

	frameLogIntervalFirstMin *= header.FPS()
	frameLogInterval *= header.FPS()

	count := 0
	t0 := time.Now()
	for {
		frame := <-spentFrames
		_, err := io.ReadFull(reader, frame)
		if err != nil {
			close(writeFrames)
			return err
		}
		totalFrames++

		if logFrameRate {
			count++
			if count == 100 {
				t1 := time.Now()
				log.Printf("%.1f Hz", float64(count)/t1.Sub(t0).Seconds())
				t0 = t1
				count = 0
			}
		}

		if totalFrames%frameLogIntervalFirstMin == 0 &&
			totalFrames <= 60*header.FPS() || totalFrames%frameLogInterval == 0 {
			log.Printf("%d frames for this connection", totalFrames)
		}

		{
			max := uint16(0)
			min := uint16((1 << 16) - 1)
			for i := 0; i < len(frame); i += 2 {
				val := binary.LittleEndian.Uint16(frame[i : i+2])
				if val < min {
					min = val
				}
				if val > max {
					max = val
				}
			}
			valRange := float64(max - min)
			j := 0
			for i := 0; i < len(frame); i += 2 {
				val := binary.LittleEndian.Uint16(frame[i : i+2])
				gray.Pix[j] = uint8((float64(val-min) / valRange) * 255.0)
				j++
			}
			draw.Draw(fb, gray.Bounds(), gray, image.ZP, draw.Src)
		}

		writeFrames <- frame
		chLen := len(writeFrames)
		if chLen > 10 && totalFrames%60 == 0 {
			log.Printf("warning: high write backlog (%d)", chLen)
		}
	}
}

func writer(inFrames <-chan []byte, conf *Config, h *headers.HeaderInfo, outFrames chan []byte) {
	builder, err := newThermalRaw(conf, time.Now(), h)
	if err != nil {
		panic(err)
	}
	changeFile := time.After(newFileInterval)
	for {
		select {
		case <-changeFile:
			builder.Close()
			builder, err = newThermalRaw(conf, time.Now(), h)
			if err != nil {
				panic(err)
			}
			changeFile = time.After(newFileInterval)
		case frame, ok := <-inFrames:
			if !ok {
				builder.Close()
				return
			}
			if err := writeFrame(builder, frame); err != nil {
				panic(err)
			}
			outFrames <- frame // Return the frame to be reused
		}
	}
}

func logConfig(conf *Config) {
	log.Printf("device name: %s", conf.DeviceName)
	log.Printf("frame input: %s", conf.FrameInput)
	log.Printf("output dir: %s", conf.OutputDir)
}
