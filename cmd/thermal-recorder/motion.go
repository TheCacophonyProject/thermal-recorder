package main

import (
	"log"

	"github.com/TheCacophonyProject/lepton3"
)

const NO_DATA = -1
const TOO_MANY_POINTS_CHANGED = -2

func NewMotionDetector(args MotionConfig) *motionDetector {

	d := new(motionDetector)
	d.flooredFrames = *NewFrameLoop(args.FrameCompareGap + 1)
	d.diffFrames = *NewFrameLoop(2)
	d.useOneFrame = args.UseOneFrameOnly
	d.framesGap = uint64(args.FrameCompareGap)
	d.deltaThresh = args.DeltaThresh
	d.countThresh = args.CountThresh
	d.tempThresh = args.TempThresh
	totalPixels := lepton3.FrameRows * lepton3.FrameCols
	d.nonzeroLimit = totalPixels * args.NonzeroMaxPercent / 100

	return d
}

type motionDetector struct {
	flooredFrames FrameLoop
	diffFrames    FrameLoop
	useOneFrame   bool
	count         uint64
	tempThresh    uint16
	deltaThresh   uint16
	countThresh   int
	nonzeroLimit  int
	framesGap     uint64
	targetX       int
	targetY       int
}

func (d *motionDetector) Detect(frame *lepton3.Frame) bool {
	movement, _ := d.pixelsChanged(frame)
	return movement
}

func (d *motionDetector) pixelsChanged(frame *lepton3.Frame) (bool, int) {
	d.count++
	log.Print("New Frame")

	processedFrame := d.flooredFrames.CurrentFrame()
	d.setFloor(frame, processedFrame)

	// we will compare with the oldest saved frame.
	compareFrame := d.flooredFrames.MoveToNextFrame()

	if d.count < d.framesGap+1 {
		return false, NO_DATA
	}

	diffFrame := d.diffFrames.CurrentFrame()
	absDiffFrames(processedFrame, compareFrame, diffFrame)

	if d.useOneFrame {
		return d.hasMotion(diffFrame, nil)
	} else {
		prevDiffFrame := d.diffFrames.MoveToNextFrame()
		return d.hasMotion(diffFrame, prevDiffFrame)
	}
}

func (d *motionDetector) setFloor(f, out *lepton3.Frame) *lepton3.Frame {
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

func (d *motionDetector) hasMotion(f1 *lepton3.Frame, f2 *lepton3.Frame) (bool, int) {
	var nonzeroCount int
	var deltaCount int
	var targetXTotal int
	var targetYTotal int
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			if f2 != nil {
				v1 := f1[y][x]
				v2 := f2[y][x]
				if (v1 > 0) && (v2 > 0) {
					nonzeroCount++
					if (v1 > d.deltaThresh) && (v2 > d.deltaThresh) {
						deltaCount++
						targetXTotal += x
						targetYTotal += y
					}
				}
			} else {
				v1 := f1[y][x]
				if v1 > 0 {
					nonzeroCount++
					if v1 > d.deltaThresh {
						deltaCount++
						targetXTotal += x
						targetYTotal += y
					}
				}
			}
		}
	}
	if deltaCount >= 1 {
		d.targetX = targetXTotal / deltaCount
		d.targetY = targetYTotal / deltaCount
	}
	// Motion detection is suppressed when over nonzeroLimit motion
	// pixels are nonzero. This is to deal with sudden jumps in the
	// readings as the camera recalibrates due to rapid temperature
	// change.
	log.Printf("deltacount %d", deltaCount)
	log.Printf("nonZeroCount %d", nonzeroCount)

	if nonzeroCount > d.nonzeroLimit {
		return false, TOO_MANY_POINTS_CHANGED
	}

	return deltaCount >= d.countThresh, deltaCount
}

func absDiffFrames(a, b, out *lepton3.Frame) *lepton3.Frame {
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
