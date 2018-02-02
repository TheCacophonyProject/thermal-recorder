package main

import (
	"github.com/TheCacophonyProject/lepton3"
)

func NewMotionDetector(args MotionConfig) *motionDetector {
	d := new(motionDetector)
	d.deltaThresh = args.DeltaThresh
	d.countThresh = args.CountThresh
	d.tempThresh = args.TempThresh
	totalPixels := lepton3.FrameRows * lepton3.FrameCols
	d.nonzeroLimit = totalPixels * args.NonzeroMaxPercent / 100
	return d
}

type motionDetector struct {
	frames       [3]*lepton3.Frame
	count        uint64
	tempThresh   uint16
	deltaThresh  uint16
	countThresh  int
	nonzeroLimit int
	tarX         int
	tarY         int
}

func (d *motionDetector) Detect(frame *lepton3.Frame) bool {
	d.count++
	d.frames[2] = d.frames[1]
	d.frames[1] = d.frames[0]
	d.frames[0] = d.setFloor(frame)
	if d.count < 3 {
		return false
	}

	// XXX this could be made more efficient by reusing delta frames and movement frames
	d1 := absDiffFrames(d.frames[0], d.frames[1])
	d2 := absDiffFrames(d.frames[1], d.frames[2])
	m := andFrames(d1, d2)
	return d.hasMotion(m)
}

func (d *motionDetector) setFloor(f *lepton3.Frame) *lepton3.Frame {
	out := new(lepton3.Frame)
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			v := f[y][x]
			if v < d.tempThresh {
				out[y][x] = d.tempThresh
			} else {
				out[y][x] = v
			}
		}
	}
	return out
}

func (d *motionDetector) hasMotion(m *lepton3.Frame) bool {
	var nonzeroCount int
	var deltaCount int
	var tarXTotal int
	var tarYTotal int
	var tarCount int
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			v := m[y][x]
			if v > 0 {
				nonzeroCount++
				if v > d.deltaThresh {
					deltaCount++
					tarXTotal += x
					tarYTotal += y
					tarCount++
				}
			}
		}
	}
	if tarCount >= 1 {
		d.tarX = tarXTotal / tarCount
		d.tarY = tarYTotal / tarCount
	}
	// Motion detection is suppressed when over nonzeroLimit motion
	// pixels are nonzero. This is to deal with sudden jumps in the
	// readings as the camera recalibrates due to rapid temperature
	// change.
	return deltaCount >= d.countThresh && nonzeroCount <= d.nonzeroLimit
}

func absDiffFrames(a, b *lepton3.Frame) *lepton3.Frame {
	out := new(lepton3.Frame)
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			out[y][x] = absDiff(a[y][x], b[y][x])
		}
	}
	return out
}

func absDiff(a, b uint16) uint16 {
	d := int32(a) - int32(b)
	if d < 0 {
		return uint16(-d)
	}
	return uint16(d)
}

func andFrames(a, b *lepton3.Frame) *lepton3.Frame {
	out := new(lepton3.Frame)
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			out[y][x] = b[y][x] & a[y][x]
		}
	}
	return out
}
