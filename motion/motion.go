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
	"math"
	"time"

	"github.com/TheCacophonyProject/lepton3"
)

// This is the period for which measurements go funny after a Flat
// Field Correction.
// TODO - this should probably be configurable (although 10s does seem right).
const ffcPeriod = 10 * time.Second

const debugLogSecs = 5
const frameBackgroundWeighting = 0.99
const weightEveryNFrames = 3

func NewMotionDetector(args MotionConfig, previewFrames int) *motionDetector {
	d := new(motionDetector)
	d.flooredFrames = *NewFrameLoop(args.FrameCompareGap + 1)
	d.diffFrames = *NewFrameLoop(2)
	d.useOneDiff = args.UseOneDiffOnly
	d.deltaThresh = args.DeltaThresh
	d.countThresh = args.CountThresh
	d.tempThresh = args.TempThresh
	d.defaultTempThresh = args.TempThresh
	d.warmerOnly = args.WarmerOnly
	d.dynamicThresh = args.DynamicThreshold
	d.start = args.EdgePixels
	d.columnStop = lepton3.FrameCols - args.EdgePixels
	d.rowStop = lepton3.FrameRows - args.EdgePixels
	d.backgroundWeight = frameBackgroundWeighting
	d.previewFrames = previewFrames
	d.numPixels = float64((d.rowStop - d.start) * (d.columnStop - d.start))

	if args.Verbose {
		d.debug = newDebugTracker()
	}

	return d
}

type motionDetector struct {
	flooredFrames     FrameLoop
	diffFrames        FrameLoop
	firstDiff         bool
	dynamicThresh     bool
	useOneDiff        bool
	tempThresh        uint16
	defaultTempThresh uint16
	deltaThresh       uint16
	countThresh       int
	warmerOnly        bool
	start             int
	rowStop           int
	columnStop        int
	count             int
	background        [lepton3.FrameRows][lepton3.FrameCols]uint16
	backgroundWeight  float32
	backgroundFrames  int
	debug             *debugTracker
	previewFrames     int
	numPixels         float64
}

func (d *motionDetector) calculateThreshold(backAverage float64) {
	d.tempThresh = uint16(math.Min(backAverage, float64(d.defaultTempThresh)))
}

func (d *motionDetector) Detect(frame *lepton3.Frame) bool {
	if d.dynamicThresh && !isAffectedByFFC(frame) {
		backAverage, changed := d.updateBackground(frame)
		if changed && d.backgroundFrames > d.previewFrames {
			d.calculateThreshold(backAverage)
			d.backgroundWeight = frameBackgroundWeighting
		} else {
			if d.count%weightEveryNFrames == 0 {
				d.backgroundWeight = d.backgroundWeight * frameBackgroundWeighting
			}
		}
		d.debug.update("thresh", int(d.tempThresh))
	}
	d.count++
	movement, deltaCount := d.pixelsChanged(frame)
	if movement {
		d.debug.update("detect", 1)
	}
	d.debug.update("delta", deltaCount)

	if d.debug != nil && d.count%(debugLogSecs*lepton3.FramesHz) == 0 {
		log.Print("motion:: " + d.debug.string("thresh:all detect:n temp:all ftemp:all diff:max delta:max ffc:n"))
		d.debug.reset()
	}
	return movement
}

func (d *motionDetector) pixelsChanged(frame *lepton3.Frame) (bool, int) {
	flooredFrame := d.flooredFrames.Current()
	d.setFloor(frame, flooredFrame)

	// we will compare with the oldest saved frame.
	compareFrame := d.flooredFrames.Oldest()
	defer d.flooredFrames.Move()

	diffFrame := d.diffFrames.Current()
	if d.warmerOnly {
		d.warmerDiffFrames(flooredFrame, compareFrame, diffFrame)
	} else {
		d.absDiffFrames(flooredFrame, compareFrame, diffFrame)
	}
	prevDiffFrame := d.diffFrames.Move()

	if !d.firstDiff {
		d.firstDiff = true
		return false, 0
	}

	if isAffectedByFFC(frame) {
		d.debug.update("ffc", 1)
		d.flooredFrames.SetAsOldest()
		d.firstDiff = false
		return false, 0
	}

	if d.useOneDiff {
		return d.hasMotion(diffFrame, nil)
	}
	return d.hasMotion(diffFrame, prevDiffFrame)
}

func isAffectedByFFC(f *lepton3.Frame) bool {
	return f.Status.TimeOn-f.Status.LastFFCTime < ffcPeriod
}

func (d *motionDetector) setFloor(f, out *lepton3.Frame) *lepton3.Frame {
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			v := f.Pix[y][x]
			d.debug.update("temp", int(v))
			if v < d.tempThresh {
				out.Pix[y][x] = d.tempThresh
			} else {
				out.Pix[y][x] = v
			}
		}
	}
	return out
}

func (d *motionDetector) CountPixelsTwoCompare(f1 *lepton3.Frame, f2 *lepton3.Frame) (deltas int) {
	var deltaCount int
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			v1 := f1.Pix[y][x]
			v2 := f2.Pix[y][x]
			if (v1 > d.deltaThresh) && (v2 > d.deltaThresh) {
				deltaCount++
			}
		}
	}
	return deltaCount
}

func (d *motionDetector) CountPixels(f1 *lepton3.Frame) (deltas int) {
	var deltaCount int
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			v := f1.Pix[y][x]
			d.debug.update("diff", int(v))
			if v > d.deltaThresh {
				deltaCount++
			}
		}
	}
	return deltaCount
}

func (d *motionDetector) hasMotion(f1 *lepton3.Frame, f2 *lepton3.Frame) (bool, int) {
	var deltaCount int
	if d.useOneDiff {
		deltaCount = d.CountPixels(f1)
	} else {
		deltaCount = d.CountPixelsTwoCompare(f1, f2)
	}
	return deltaCount >= d.countThresh, deltaCount
}

func (d *motionDetector) absDiffFrames(a, b, out *lepton3.Frame) *lepton3.Frame {
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			out.Pix[y][x] = absDiff(a.Pix[y][x], b.Pix[y][x])
		}
	}
	return out
}

func (d *motionDetector) warmerDiffFrames(a, b, out *lepton3.Frame) *lepton3.Frame {
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			va := a.Pix[y][x]
			d.debug.update("ftemp", int(va))
			out.Pix[y][x] = warmerDiff(va, b.Pix[y][x])
		}
	}
	return out
}

func (d *motionDetector) updateBackground(new_frame *lepton3.Frame) (float64, bool) {
	d.backgroundFrames++
	if d.backgroundFrames == 1 {
		d.background = new_frame.Pix
		return 0, true
	}

	var changed bool = false
	var average float64 = 0
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			if uint16(float32(new_frame.Pix[y][x])*d.backgroundWeight) < d.background[y][x] {
				d.background[y][x] = new_frame.Pix[y][x]
				changed = true
			}
			average = average + float64(d.background[y][x])/d.numPixels
		}
	}
	return average, changed
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
