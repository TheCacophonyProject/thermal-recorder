package main

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
	tfm.processor.processFrame(frame)
}

func (tfm *TestFrameMaker) makeFrame() *lepton3.Frame {
	frame := new(lepton3.Frame)

	if tfm.BackgroundVal != 0 {
		for y := 0; y < lepton3.FrameRows; y++ {
			for x := 0; x < lepton3.FrameCols; x++ {
				frame[y][x] = uint16(tfm.BackgroundVal)
			}
		}
	}

	frame[0][0] = tfm.frameCounter
	tfm.frameCounter++
	return frame
}

func (tfm *TestFrameMaker) makeFrameWithBrightSport(position int) *lepton3.Frame {
	frame := tfm.makeFrame()

	brightness16 := uint16(tfm.BackgroundVal + tfm.BrightSpotVal)

	frame[position][position] = brightness16
	frame[position+1][position] = brightness16
	frame[position+2][position] = brightness16
	frame[position][position+1] = brightness16
	frame[position+1][position+1] = brightness16
	frame[position+2][position+1] = brightness16
	frame[position][position+2] = brightness16
	frame[position+1][position+2] = brightness16
	frame[position+2][position+2] = brightness16

	return frame
}
