// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	"periph.io/x/periph/conn/gpio"
	"periph.io/x/periph/conn/gpio/gpioreg"
	"periph.io/x/periph/host"

	"github.com/TheCacophonyProject/thermal_recorder/output"
)

// XXX most of this should come from a configuration file
const (
	spiSpeed  = 30000000
	powerPin  = "GPIO23"
	directory = "./r"
)

func main() {
	err := runMain()
	if err != nil {
		log.Fatal(err)
	}
}

func runMain() error {
	_, err := host.Init()
	if err != nil {
		return err
	}

	if err := powerupCamera(powerPin); err != nil {
		return err
	}

	f, err := os.Create("out.cptv")
	if err != nil {
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	fw := output.NewWriter(w)
	defer fw.Close()

	hf := output.NewFields()
	t0 := time.Now()
	hf.Timestamp(output.Timestamp, t0)
	hf.Uint32(output.XResolution, lepton3.FrameCols)
	hf.Uint32(output.YResolution, lepton3.FrameRows)
	hf.Uint8(output.Compression, 1)
	fw.WriteHeader(hf)

	camera := lepton3.New(spiSpeed)
	err = camera.Open()
	if err != nil {
		return err
	}
	defer camera.Close()
	camera.SetLogFunc(func(t string) { log.Printf(t) })

	compressor := output.NewCompressor(lepton3.FrameCols, lepton3.FrameRows)
	prevFrame := new(lepton3.Frame)
	frame := new(lepton3.Frame)
	for i := 0; i < 10; i++ {
		err := camera.NextFrame(frame)
		if err != nil {
			return err
		}
		dt := uint64(time.Since(t0))

		bitWidth, cFrame := compressor.Next(prevFrame, frame)

		ff := output.NewFields()
		ff.Uint32(output.Offset, uint32(dt/1000))
		ff.Uint8(output.BitWidth, uint8(bitWidth))
		ff.Uint32(output.FrameSize, uint32(len(cFrame)))
		fw.WriteFrame(ff, cFrame)

		frame, prevFrame = prevFrame, frame
	}

	return nil
}

func flattenFrame(f *lepton3.Frame) []byte {
	out := new(bytes.Buffer)
	out.Grow(2 * lepton3.FrameRows * lepton3.FrameCols)
	for y := 0; y < len(f); y++ {
		binary.Write(out, binary.LittleEndian, f[y])
	}
	return out.Bytes()
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
