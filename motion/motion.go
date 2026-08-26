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

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
)

// This is the period for which measurements go funny after a Flat
// Field Correction.
// TODO - this should probably be configurable (although 10s does seem right).
const ffcPeriod = 10 * time.Second

const debugLogSecs = 5

func NewMotionDetector(args config.ThermalMotion, previewFrames int, camera cptvframe.CameraSpec) *motionDetector {
	d := new(motionDetector)
	d.flooredFrames = *NewFrameLoop(args.FrameCompareGap+1, camera)
	d.diffFrames = *NewFrameLoop(2, camera)
	d.useOneDiff = args.UseOneDiffOnly
	d.deltaThresh = args.DeltaThresh
	d.countThresh = args.CountThresh
	d.tempThresh = args.TempThresh
	d.tempThreshMin = args.TempThreshMin
	d.tempThreshMax = args.TempThreshMax
	d.warmerOnly = args.WarmerOnly
	d.dynamicThresh = args.DynamicThreshold
	d.start = args.EdgePixels
	d.columnStop = camera.ResX() - args.EdgePixels
	d.rowStop = camera.ResY() - args.EdgePixels

	d.previewFrames = previewFrames
	d.numPixels = float64((d.rowStop - d.start) * (d.columnStop - d.start))
	d.framesHz = camera.FPS()
	if args.Verbose {
		d.debug = newDebugTracker()
	}
	d.background = cptvframe.NewFrame(camera)
	d.background.Status.BackgroundFrame = true
	d.backgroundWeight = make([][]float32, camera.ResY())
	for i := range d.backgroundWeight {
		d.backgroundWeight[i] = make([]float32, camera.ResX())
	}

	return d
}

type motionDetector struct {
	flooredFrames    FrameLoop
	diffFrames       FrameLoop
	firstDiff        bool
	dynamicThresh    bool
	useOneDiff       bool
	tempThresh       uint16
	tempThreshMax    uint16
	tempThreshMin    uint16
	deltaThresh      uint16
	countThresh      int
	warmerOnly       bool
	start            int
	rowStop          int
	columnStop       int
	count            int
	background       *cptvframe.Frame
	backgroundWeight [][]float32
	backgroundFrames int
	debug            *debugTracker
	previewFrames    int
	numPixels        float64
	affectedByFCC    bool
	framesHz         int
}

func (d *motionDetector) Reset(camera cptvframe.CameraSpec) {
	d.backgroundFrames = 0
	d.count = 0
	d.flooredFrames.Reset()
	d.diffFrames.Reset()
}

func (d *motionDetector) calculateThreshold(backAverage float64) {
	if d.tempThreshMin != 0 {
		d.tempThresh = uint16(math.Max(backAverage, float64(d.tempThreshMin)))
	} else {
		d.tempThresh = uint16(backAverage)
	}
	if d.tempThreshMax != 0 {
		d.tempThresh = uint16(math.Min(backAverage, float64(d.tempThreshMax)))
	}
}

func (d *motionDetector) Detect(frame *cptvframe.Frame) bool {
	prevFFC := d.affectedByFCC
	d.affectedByFCC = isAffectedByFFC(frame)
	if d.dynamicThresh && !d.affectedByFCC {
		backAverage, changed := d.updateBackground(frame)
		if changed && d.backgroundFrames > d.previewFrames {
			d.calculateThreshold(backAverage)
		}
		d.debug.update("thresh", int(d.tempThresh))
	}
	d.count++
	movement, deltaCount := d.pixelsChanged(frame, prevFFC)
	if movement {
		d.debug.update("detect", 1)
	}
	d.debug.update("delta", deltaCount)

	if d.debug != nil && d.count%(debugLogSecs*d.framesHz) == 0 {
		log.Print("motion:: " + d.debug.string("thresh:all detect:n temp:all ftemp:all diff:max delta:max ffc:n"))
		d.debug.reset()
	}
	return movement
}

func (d *motionDetector) pixelsChanged(frame *cptvframe.Frame, prevFFC bool) (bool, int) {
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

	if isAffectedByFFC(frame) || prevFFC {
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

func isAffectedByFFC(f *cptvframe.Frame) bool {
	return f.Status.TimeOn-f.Status.LastFFCTime < ffcPeriod
}

func (d *motionDetector) setFloor(f, out *cptvframe.Frame) *cptvframe.Frame {
	out.Copy(f)
	return out
}

func (d *motionDetector) CountPixelsTwoCompare(f1 *cptvframe.Frame, f2 *cptvframe.Frame) (deltas int) {
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

func (d *motionDetector) CountPixels(f1 *cptvframe.Frame) (deltas int) {
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

func (d *motionDetector) hasMotion(f1 *cptvframe.Frame, f2 *cptvframe.Frame) (bool, int) {
	var deltaCount int
	if d.useOneDiff {
		deltaCount = d.CountPixels(f1)
	} else {
		deltaCount = d.CountPixelsTwoCompare(f1, f2)
	}
	return deltaCount >= d.countThresh, deltaCount
}

func (d *motionDetector) absDiffFrames(a, b, out *cptvframe.Frame) *cptvframe.Frame {
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			va := a.Pix[y][x]
			if va < d.tempThresh {
				va = d.tempThresh
			}
			vb := b.Pix[y][x]
			if vb < d.tempThresh {
				vb = d.tempThresh
			}
			out.Pix[y][x] = absDiff(va, vb)
		}
	}
	return out
}

func (d *motionDetector) warmerDiffFrames(a, b, out *cptvframe.Frame) *cptvframe.Frame {
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			va := a.Pix[y][x]
			d.debug.update("ftemp", int(va))
			if va < d.tempThresh {
				va = d.tempThresh
			}
			vb := b.Pix[y][x]
			if vb < d.tempThresh {
				vb = d.tempThresh
			}
			out.Pix[y][x] = warmerDiff(va, vb)
		}
	}
	return out
}

func (d *motionDetector) updateBackground(new_frame *cptvframe.Frame) (float64, bool) {
	d.backgroundFrames++
	if d.backgroundFrames == 1 {
		for y := d.start; y < d.rowStop; y++ {
			copy(d.background.Pix[y][d.start:d.columnStop], new_frame.Pix[y][d.start:d.columnStop])
			for x := 0; x < d.start; x++ {
				d.background.Pix[y][x] = new_frame.Pix[y][d.start]
				d.background.Pix[y][d.columnStop+x] = new_frame.Pix[y][d.columnStop-1]
			}
		}
		for y := 0; y < d.start; y++ {
			copy(d.background.Pix[y], d.background.Pix[d.start])
			copy(d.background.Pix[d.rowStop+y], d.background.Pix[d.rowStop-1])
		}

		return 0, true
	}

	var changed bool = false
	var average float64 = 0
	for y := d.start; y < d.rowStop; y++ {
		for x := d.start; x < d.columnStop; x++ {
			weight := d.backgroundWeight[y][x]
			if (float32(new_frame.Pix[y][x]) - weight) < float32(d.background.Pix[y][x]) {
				d.background.Pix[y][x] = new_frame.Pix[y][x]
				d.backgroundWeight[y][x] = 0
				changed = true
			} else {
				weight += 0.1
				if weight > math.MaxFloat32 {
					weight = math.MaxFloat32
				}
				d.backgroundWeight[y][x] = weight
			}
			average = average + float64(d.background.Pix[y][x])/d.numPixels
			for x := 0; x < d.start; x++ {
				// copy valid pixels into edge pixels
				d.background.Pix[y][x] = d.background.Pix[y][d.start]
				d.background.Pix[y][d.columnStop+x] = d.background.Pix[y][d.columnStop-1]
			}
		}
	}
	// copy valid pixels into edge pixels
	for y := 0; y < d.start; y++ {
		copy(d.background.Pix[y], d.background.Pix[d.start])
		copy(d.background.Pix[d.rowStop+y], d.background.Pix[d.rowStop-1])
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
