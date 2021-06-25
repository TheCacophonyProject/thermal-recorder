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

package motion

import (
	"time"

	"github.com/TheCacophonyProject/go-cptv/cptvframe"
)

// TODO - this is very similar to frameGen in motion_test.go. Merge them.
type TestFrameMaker struct {
	frameCounter       uint16
	processor          *MotionProcessor
	BackgroundVal      int
	BrightSpotVal      int
	brightSpotPosition int
	now                time.Duration
	camera             cptvframe.CameraSpec
}

func MakeTestFrameMaker(motionProcessor *MotionProcessor, camera cptvframe.CameraSpec) *TestFrameMaker {
	return &TestFrameMaker{
		processor:     motionProcessor,
		BackgroundVal: 3300,
		BrightSpotVal: 100,
		now:           time.Minute,
		camera:        camera,
	}
}

func (tfm *TestFrameMaker) AddBackgroundFrames(frames int) *TestFrameMaker {
	for i := 0; i < frames; i++ {
		frame := tfm.makeFrame()
		tfm.PlayFrame(frame)
	}
	return tfm
}

func (tfm *TestFrameMaker) AddMovingDotFrames(frames int) *TestFrameMaker {
	for i := 0; i < frames; i++ {
		tfm.brightSpotPosition += 3
		frame := tfm.makeFrameWithBrightSport(tfm.brightSpotPosition)
		tfm.PlayFrame(frame)
	}
	return tfm
}

func (tfm *TestFrameMaker) PlayFrame(frame *cptvframe.Frame) {
	tfm.processor.ProcessFrame(frame)
}

func (tfm *TestFrameMaker) makeFrame() *cptvframe.Frame {
	frame := cptvframe.NewFrame(tfm.camera)
	frame.Status.TimeOn = tfm.now
	tfm.now += frameInterval

	if tfm.BackgroundVal != 0 {
		for y, row := range frame.Pix {
			for x := range row {
				frame.Pix[y][x] = uint16(tfm.BackgroundVal)
			}
		}
	}

	frame.Pix[0][0] = tfm.frameCounter
	tfm.frameCounter++
	return frame
}

func (tfm *TestFrameMaker) makeFrameWithBrightSport(position int) *cptvframe.Frame {
	frame := tfm.makeFrame()

	brightness16 := uint16(tfm.BackgroundVal + tfm.BrightSpotVal)

	frame.Pix[position][position] = brightness16
	frame.Pix[position+1][position] = brightness16
	frame.Pix[position+2][position] = brightness16
	frame.Pix[position][position+1] = brightness16
	frame.Pix[position+1][position+1] = brightness16
	frame.Pix[position+2][position+1] = brightness16
	frame.Pix[position][position+2] = brightness16
	frame.Pix[position+1][position+2] = brightness16
	frame.Pix[position+2][position+2] = brightness16

	return frame
}
