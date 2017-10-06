package main

import (
	"github.com/TheCacophonyProject/lepton3"
)

type motionDetector struct {
	frames      [3]*lepton3.Frame
	count       uint64
	deltaThresh uint16
	countThresh uint16
	tempThresh  uint16
}

func NewMotionDetector(deltaThresh, countThresh, tempThresh uint16) *motionDetector {
	d := new(motionDetector)
	d.deltaThresh = deltaThresh
	d.countThresh = countThresh
	d.tempThresh = tempThresh
	return d
}

func (d *motionDetector) Detect(frame *lepton3.Frame) bool {
	d.count++
	d.frames[2] = d.frames[1]
	d.frames[1] = d.frames[0]
	d.frames[0] = d.stripLow(frame)
	if d.count < 3 {
		return false
	}

	// XXX this could be made more efficient by reusing delta frames and movement frames
	d1 := absDiffFrames(d.frames[0], d.frames[1])
	d2 := absDiffFrames(d.frames[1], d.frames[2])
	m := andFrames(d1, d2)
	return d.hasMotion(m)
}

func (d *motionDetector) stripLow(f *lepton3.Frame) *lepton3.Frame {
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

func (d *motionDetector) hasMotion(f *lepton3.Frame) bool {
	var count uint16 = 0
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			if f[y][x] > d.deltaThresh {
				count++
			}
			if count >= d.countThresh {
				return true
			}
		}
	}
	return false
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
