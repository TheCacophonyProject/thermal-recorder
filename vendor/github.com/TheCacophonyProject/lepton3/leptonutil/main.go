// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

// This implements a simple utility for pulling frames off the Lepton
// 3 thermal camera.

package main

import (
	"errors"
	"fmt"
	"log"
	"path/filepath"

	arg "github.com/alexflint/go-arg"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/TheCacophonyProject/lepton3"
)

type Options struct {
	Frames    int    `arg:"-f,help:number of frames to collect (default=all)"`
	Speed     int64  `arg:"-s,help:SPI speed in MHz"`
	Directory string `arg:"-d,help:Directory to write output files"`
	PowerPin  string `arg:"-p,help:Optional pin to set to power on camera"`
	Output    string `arg:"positional,required,help:png or none"`
}

func procCommandLine() Options {
	opts := Options{}
	opts.Speed = 30
	opts.Directory = "."
	arg.MustParse(&opts)
	if opts.Output != "png" && opts.Output != "none" {
		log.Fatalf("invalid output type: %q", opts.Output)
	}
	opts.Speed *= 1000000 // convert to Hz
	return opts
}

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	opts := procCommandLine()

	_, err := host.Init()
	if err != nil {
		return err
	}

	if opts.PowerPin != "" {
		powerPin := gpioreg.ByName(opts.PowerPin)
		if powerPin == nil {
			return errors.New("unable to load power pin")
		}
		if err := powerPin.Out(gpio.High); err != nil {
			return errors.New("failed to set camera power pin")
		}
	}

	camera := lepton3.New(opts.Speed)
	if err := camera.SetRadiometry(true); err != nil {
		return err
	}

	err = camera.Open()
	if err != nil {
		return err
	}
	defer camera.Close()

	camera.SetLogFunc(func(t string) {
		log.Printf(t)
	})

	frame := new(lepton3.Frame)
	i := 0
	for {
		err := camera.NextFrame(frame)
		if err != nil {
			return err
		}
		fmt.Printf(".")

		if opts.Output == "png" {
			filename := filepath.Join(opts.Directory, fmt.Sprintf("%05d.png", i))
			err := dumpToPNG(filename, frame)
			if err != nil {
				return err
			}
		}

		i++
		if opts.Frames > 0 && i >= opts.Frames {
			break
		}
	}
	fmt.Println()

	return nil
}
