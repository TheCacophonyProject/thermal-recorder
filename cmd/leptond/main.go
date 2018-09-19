// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

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

	conn, err := connectToFrameOutput(conf.FrameOutput)
	if err != nil {
		return errors.New("error: connecting to frame output socket failed")
	}
	conn.Close()

	log.Print("host initialisation")
	if _, err := host.Init(); err != nil {
		return err
	}

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

		err := runCamera(conf, camera)
		if err != nil {
			if _, isNextFrameErr := err.(*nextFrameErr); !isNextFrameErr {
				return err
			}
			log.Printf("recording error: %v", err)
		}

		log.Print("closing camera")
		camera.Close()

		err = cycleCameraPower(conf.PowerPin)
		if err != nil {
			return err
		}
	}
}

func runCamera(conf *Config, camera *lepton3.Lepton3) error {
	log.Print("dialing frame output socket")
	conn, err := connectToFrameOutput(conf.FrameOutput)
	if err != nil {
		return err
	}
	defer conn.Close()
	conn.SetWriteBuffer(lepton3.FrameCols * lepton3.FrameRows * 2 * 20)

	log.Print("reading frames")
	frame := new(lepton3.RawFrame)
	notifyCount := 0
	for {
		if err := camera.NextFrame(frame); err != nil {
			return &nextFrameErr{err}
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

func connectToFrameOutput(path string) (*net.UnixConn, error) {
	return net.DialUnix("unixpacket", nil, &net.UnixAddr{
		Net:  "unixgram",
		Name: path,
	})
}
