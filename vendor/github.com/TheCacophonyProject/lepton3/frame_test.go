// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package lepton3

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFrameCopy(t *testing.T) {
	frame := new(Frame)
	frame[0][0] = 1
	frame[9][7] = 2
	frame[FrameRows-1][0] = 3
	frame[0][FrameCols-1] = 4
	frame[FrameRows-1][FrameCols-1] = 5

	frame2 := new(Frame)

	frame2.Copy(frame)
	assert.Equal(t, 1, int(frame[0][0]))
	assert.Equal(t, 2, int(frame[9][7]))
	assert.Equal(t, 3, int(frame[FrameRows-1][0]))
	assert.Equal(t, 4, int(frame[0][FrameCols-1]))
	assert.Equal(t, 5, int(frame[FrameRows-1][FrameCols-1]))
}
