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
	"net"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	arg "github.com/alexflint/go-arg"
	"github.com/coreos/go-systemd/daemon"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"
)

const (
	framesHz = 9 // approx

	frameLogIntervalFirstMin = 15 * framesHz
	frameLogInterval         = 60 * 5 * framesHz

	framesPerSdNotify = 5 * framesHz

	// ffcInterval is how often to force Flat Field Correction (FFC)
	// By default, the camera will do this every 5 minutes but we want
	// to be in control so do it every 4.5 minutes.
	ffcInterval = ((5 * 60) - 30) * framesHz
)

var version = "<not set>"

type Args struct {
	ConfigFile string `arg:"-c,--config" help:"path to configuration file"`
	Quick      bool   `arg:"-q,--quick" help:"don't cycle camera power on startup"`
	Timestamps bool   `arg:"-t,--timestamps" help:"include timestamps in log output"`
}

func (Args) Version() string {
	return version
}

func procArgs() Args {
	var args Args
	args.ConfigFile = "/etc/leptond.yaml"
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
	if !args.Timestamps {
		log.SetFlags(0) // Removes default timestamp flag
	}

	log.Printf("version: %s", version)
	conf, err := ParseConfigFile(args.ConfigFile)
	if err != nil {
		return err
	}
	logConfig(conf)

	log.Print("dialing frame output socket")
	conn, err := dialOutput(conf.FrameOutput)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Print("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

	// Camera power is cycled *after* an output connection has been
	// established to avoid continually rebooting the camera if the
	// downstream process is not running.
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

		log.Print("opening camera")
		if err := camera.Open(); err != nil {
			return err
		}

		log.Print("enabling radiometry")
		if err := camera.SetRadiometry(true); err != nil {
			return err
		}

		err := runCamera(conf, camera, conn)

		// Drop output connection to indicate that the frames have
		// stopped.
		conn.Close()

		if err == nil {
			// Run FFC
			log.Print("starting FFC")
			camera.RunFFC()
			camera.Close()
			log.Print("waiting for FFC to complete")
			time.Sleep(9 * time.Second)
			log.Print("FFC done")
		} else {
			// Failed I/O with camera.
			if _, isNextFrameErr := err.(*nextFrameErr); !isNextFrameErr {
				return err
			}
			log.Printf("recording error: %v", err)
			camera.Close()

			if err := cycleCameraPower(conf.PowerPin); err != nil {
				return err
			}
		}

		conn, err = dialOutput(conf.FrameOutput)
		if err != nil {
			return err
		}
	}
}

func dialOutput(name string) (*net.UnixConn, error) {
	log.Print("dialing frame output socket")
	conn, err := net.DialUnix("unixpacket", nil, &net.UnixAddr{
		Net:  "unixgram",
		Name: name,
	})
	if err != nil {
		return nil, errors.New("error: connecting to frame output socket failed")
	}
	return conn, err
}

func runCamera(conf *Config, camera *lepton3.Lepton3, conn *net.UnixConn) error {
	conn.SetWriteBuffer(lepton3.FrameCols * lepton3.FrameRows * 2 * 20)

	log.Print("reading frames")
	frame := new(lepton3.RawFrame)

	notifyCount := 0
	frameCount := 0
	for {
		if err := camera.NextFrame(frame); err != nil {
			return &nextFrameErr{err}
		}

		frameCount++
		if frameCount >= ffcInterval {
			return nil
		}

		if notifyCount++; notifyCount >= framesPerSdNotify {
			daemon.SdNotify(false, "WATCHDOG=1")
			notifyCount = 0
		}

		if _, err := conn.Write(frame[:]); err != nil {
			return err
		}
	}
}

func logConfig(conf *Config) {
	log.Printf("SPI speed: %d", conf.SPISpeed)
	log.Printf("power pin: %s", conf.PowerPin)
	log.Printf("frame output: %s", conf.FrameOutput)
}

func cycleCameraPower(pinName string) error {
	if pinName == "" {
		return nil
	}

	pin := gpioreg.ByName(pinName)

	log.Print("turning camera power off")
	if err := pin.Out(gpio.Low); err != nil {
		return fmt.Errorf("failed to set camera power pin low: %v", err)
	}
	time.Sleep(2 * time.Second)

	log.Print("turning camera power on")
	if err := pin.Out(gpio.High); err != nil {
		return fmt.Errorf("failed to set camera power pin high: %v", err)
	}

	log.Print("waiting for camera startup")
	time.Sleep(8 * time.Second)
	log.Print("camera should be ready")
	return nil
}
