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
	"log"

	"github.com/TheCacophonyProject/lepton3"
)

const NO_DATA = -1
const TOO_MANY_POINTS_CHANGED = -2

func NewMotionDetector(args MotionConfig) *motionDetector {

	d := new(motionDetector)
	d.flooredFrames = *NewFrameLoop(args.FrameCompareGap + 1)
	d.diffFrames = *NewFrameLoop(2)
	d.useOneDiff = args.UseOneDiffOnly
	d.framesGap = uint64(args.FrameCompareGap)
	d.deltaThresh = args.DeltaThresh
	d.countThresh = args.CountThresh
	d.tempThresh = args.TempThresh
	totalPixels := lepton3.FrameRows * lepton3.FrameCols
	d.nonzeroLimit = totalPixels * args.NonzeroMaxPercent / 100
	d.verbose = args.Verbose
	d.warmerOnly = args.WarmerOnly

	return d
}

type motionDetector struct {
	flooredFrames FrameLoop
	diffFrames    FrameLoop
	firstDiff     bool
	useOneDiff    bool
	tempThresh    uint16
	deltaThresh   uint16
	countThresh   int
	nonzeroLimit  int
	framesGap     uint64
	verbose       bool
	warmerOnly    bool
}

func (d *motionDetector) Detect(frame *lepton3.Frame) bool {
	movement, _ := d.pixelsChanged(frame)
	return movement
}

func (d *motionDetector) pixelsChanged(frame *lepton3.Frame) (bool, int) {

	processedFrame := d.flooredFrames.Current()
	d.setFloor(frame, processedFrame)

	// we will compare with the oldest saved frame.
	compareFrame := d.flooredFrames.Oldest()
	defer d.flooredFrames.Move()

	diffFrame := d.diffFrames.Current()
	if d.warmerOnly {
		warmerDiffFrames(processedFrame, compareFrame, diffFrame)
	} else {
		absDiffFrames(processedFrame, compareFrame, diffFrame)
	}
	prevDiffFrame := d.diffFrames.Move()

	if !d.firstDiff {
		d.firstDiff = true
		return false, NO_DATA
	}

	if d.useOneDiff {
		return d.hasMotion(diffFrame, nil)
	} else {
		return d.hasMotion(diffFrame, prevDiffFrame)
	}
}

func (d *motionDetector) setFloor(f, out *lepton3.Frame) *lepton3.Frame {
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			v := f.Pix[y][x]
			if v < d.tempThresh {
				out.Pix[y][x] = d.tempThresh
			} else {
				out.Pix[y][x] = v
			}
		}
	}
	return out
}

func (d *motionDetector) CountPixelsTwoCompare(f1 *lepton3.Frame, f2 *lepton3.Frame) (nonZeros, deltas int) {
	var nonzeroCount int
	var deltaCount int
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			v1 := f1.Pix[y][x]
			v2 := f2.Pix[y][x]
			if (v1 > 0) || (v2 > 0) {
				nonzeroCount++
				if (v1 > d.deltaThresh) && (v2 > d.deltaThresh) {
					deltaCount++
				}
			}
		}
	}
	return nonzeroCount, deltaCount
}

func (d *motionDetector) CountPixels(f1 *lepton3.Frame) (nonZeros, deltas int) {
	var nonzeroCount int
	var deltaCount int
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			v1 := f1.Pix[y][x]
			if v1 > 0 {
				nonzeroCount++
				if v1 > d.deltaThresh {
					if d.verbose {
						log.Printf("Motion (%d, %d) = %d", x, y, v1)
					}
					deltaCount++
				}
			}
		}
	}
	return nonzeroCount, deltaCount
}

func (d *motionDetector) hasMotion(f1 *lepton3.Frame, f2 *lepton3.Frame) (bool, int) {
	var nonzeroCount int
	var deltaCount int
	if d.useOneDiff {
		nonzeroCount, deltaCount = d.CountPixels(f1)
	} else {
		nonzeroCount, deltaCount = d.CountPixelsTwoCompare(f1, f2)
	}

	// Motion detection is suppressed when over nonzeroLimit motion
	// pixels are nonzero. This is to deal with sudden jumps in the
	// readings as the camera recalibrates due to rapid temperature
	// change.

	if nonzeroCount > d.nonzeroLimit {
		log.Printf("Motion detector - too many points changed, probably a recalculation")
		d.flooredFrames.SetAsOldest()
		d.firstDiff = false
		return false, TOO_MANY_POINTS_CHANGED
	}

	if deltaCount > 0 && d.verbose {
		log.Printf("deltaCount %d", deltaCount)
	}
	return deltaCount >= d.countThresh, deltaCount
}

func absDiffFrames(a, b, out *lepton3.Frame) *lepton3.Frame {
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			out.Pix[y][x] = absDiff(a.Pix[y][x], b.Pix[y][x])
		}
	}
	return out
}

func warmerDiffFrames(a, b, out *lepton3.Frame) *lepton3.Frame {
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			out.Pix[y][x] = warmerDiff(a.Pix[y][x], b.Pix[y][x])
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

func warmerDiff(a, b uint16) uint16 {
	d := int32(a) - int32(b)

	if d < 0 {
		return 0
	}
	return uint16(d)
}
