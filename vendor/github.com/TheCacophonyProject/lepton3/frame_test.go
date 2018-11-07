// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package lepton3

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFrameCopy(t *testing.T) {
	frame := new(Frame)

	// Pixel values.
	frame.Pix[0][0] = 1
	frame.Pix[9][7] = 2
	frame.Pix[FrameRows-1][0] = 3
	frame.Pix[0][FrameCols-1] = 4
	frame.Pix[FrameRows-1][FrameCols-1] = 5
	// Status values.
	frame.Status.TimeOn = 10 * time.Second
	frame.Status.FrameCount = 123
	frame.Status.TempC = 23.1

	frame2 := new(Frame)
	frame2.Copy(frame)

	assert.Equal(t, 1, int(frame2.Pix[0][0]))
	assert.Equal(t, 2, int(frame2.Pix[9][7]))
	assert.Equal(t, 3, int(frame2.Pix[FrameRows-1][0]))
	assert.Equal(t, 4, int(frame2.Pix[0][FrameCols-1]))
	assert.Equal(t, 5, int(frame2.Pix[FrameRows-1][FrameCols-1]))
	assert.Equal(t, frame.Status, frame2.Status)
}
