// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"

	"github.com/TheCacophonyProject/lepton3"
)

func dumpToPNG(path string, frame *lepton3.Frame) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	w := bufio.NewWriter(f)
	defer func() {
		w.Flush()
		f.Close()
	}()
	return png.Encode(w, reduce(frame))
}

var dst = image.NewGray16(image.Rect(0, 0, lepton3.FrameCols, lepton3.FrameRows))

func reduce(src *lepton3.Frame) *image.Gray16 {
	minVal := uint16(math.MaxUint16)
	maxVal := uint16(0)
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			i := src.Pix[y][x]
			if i > maxVal {
				maxVal = i
			}
			if i < minVal {
				minVal = i
			}
		}
	}

	var norm = math.MaxUint16 / (maxVal - minVal)
	for y, row := range src.Pix {
		for x, val := range row {
			dst.SetGray16(x, y, color.Gray16{Y: (val - minVal) * norm})
		}
	}
	return dst
}
