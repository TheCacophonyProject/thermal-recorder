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
	"github.com/TheCacophonyProject/lepton3"
)

type TestFrameMaker struct {
	frameCounter       uint16
	processor          *MotionProcessor
	BackgroundVal      int
	BrightSpotVal      int
	brightSpotPosition int
}

func MakeTestFrameMaker(motionProcessor *MotionProcessor) *TestFrameMaker {
	return &TestFrameMaker{
		processor:     motionProcessor,
		BackgroundVal: 3300,
		BrightSpotVal: 100,
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

func (tfm *TestFrameMaker) PlayFrame(frame *lepton3.Frame) {
	tfm.processor.ProcessFrame(frame)
}

func (tfm *TestFrameMaker) makeFrame() *lepton3.Frame {
	frame := new(lepton3.Frame)

	if tfm.BackgroundVal != 0 {
		for y := 0; y < lepton3.FrameRows; y++ {
			for x := 0; x < lepton3.FrameCols; x++ {
				frame.Pix[y][x] = uint16(tfm.BackgroundVal)
			}
		}
	}

	frame.Pix[0][0] = tfm.frameCounter
	tfm.frameCounter++
	return frame
}

func (tfm *TestFrameMaker) makeFrameWithBrightSport(position int) *lepton3.Frame {
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
