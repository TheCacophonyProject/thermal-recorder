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
	"fmt"
	"log"
	"testing"
	"time"

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
)

type TestCamera struct {
}

func (cam *TestCamera) ResX() int {
	return 160
}
func (cam *TestCamera) ResY() int {
	return 120
}
func (cam *TestCamera) FPS() int {
	return 9
}

func TestRevertsToUsingSmallerFrameIntervalWhenNotEnoughFrames_OneFrame(t *testing.T) {
	camera := new(TestCamera)
	config := defaultMotionParams()
	config.UseOneDiffOnly = true
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	detects, pixels := newFrameGen(detector, camera).Movement(5)
	assert.Equal(t, []bool{false, true, true, true, true}, detects)
	assert.Equal(t, []int{0, 9, 9, 9, 18}, pixels)
}

func TestNoMotionDetectedIfNothingHasChanged(t *testing.T) {
	camera := new(TestCamera)
	config := defaultMotionParams()
	config.UseOneDiffOnly = true
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	detects, pixels := newFrameGen(detector, camera).NoMovement(5)
	assertAllFalse(t, detects)
	assertAllZero(t, pixels)
}

func TestIfUsingTwoFramesItOnlyCountsWhereBothFramesHaveChanged(t *testing.T) {
	camera := new(TestCamera)
	config := defaultMotionParams()
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	detects, pixels := newFrameGen(detector, camera).Movement(6)
	assert.Equal(t, []bool{false, false, false, false, false, true}, detects)
	assert.Equal(t, []int{0, 0, 4, 4, 5, 9}, pixels)
}

func TestChangeCountThresh(t *testing.T) {
	camera := new(TestCamera)
	config := defaultMotionParams()
	config.CountThresh = 4
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	detects, pixels := newFrameGen(detector, camera).Movement(6)
	assert.Equal(t, []bool{false, false, true, true, true, true}, detects)
	assert.Equal(t, []int{0, 0, 4, 4, 5, 9}, pixels)
}

func TestIgnoresEdgePixel(t *testing.T) {
	camera := new(TestCamera)

	config := defaultMotionParams()
	config.EdgePixels = 1
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	detects, pixels := newFrameGen(detector, camera).MovementInColumn(0, 4)
	assert.Equal(t, []bool{false, false, false, false}, detects)
	assert.Equal(t, []int{0, 0, 0, 0}, pixels)

	detects, pixels = newFrameGen(detector, camera).MovementInColumn(lepton3.FrameCols-1, 4)
	assert.Equal(t, []bool{false, false, false, false}, detects)
	assert.Equal(t, []int{0, 0, 0, 0}, pixels)

	detects, pixels = newFrameGen(detector, camera).MovementInRow(0, 4)
	assert.Equal(t, []bool{false, false, false, false}, detects)
	assert.Equal(t, []int{0, 0, 0, 0}, pixels)

	detects, pixels = newFrameGen(detector, camera).MovementInRow(lepton3.FrameRows-1, 4)
	assert.Equal(t, []bool{false, false, false, false}, detects)
	assert.Equal(t, []int{0, 0, 0, 0}, pixels)
}

func TestDetectsAfterEdgePixel(t *testing.T) {
	camera := new(TestCamera)

	config := defaultMotionParams()
	config.EdgePixels = 1
	config.WarmerOnly = true
	config.CountThresh = 4
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	detects, pixels := newFrameGen(detector, camera).MovementInColumn(1, 4)
	assert.Equal(t, []bool{false, false, true, true}, detects)
	assert.Equal(t, []int{0, 0, 6, 6}, pixels)

	detects, pixels = newFrameGen(detector, camera).MovementInColumn(lepton3.FrameCols-2, 4)
	assert.Equal(t, []bool{false, true, true, true}, detects)
	assert.Equal(t, []int{0, 6, 6, 6}, pixels)

	detects, pixels = newFrameGen(detector, camera).MovementInRow(1, 4)
	assert.Equal(t, []bool{false, true, true, true}, detects)
	assert.Equal(t, []int{0, 6, 6, 6}, pixels)

	detects, pixels = newFrameGen(detector, camera).MovementInRow(lepton3.FrameRows-2, 4)
	assert.Equal(t, []bool{false, true, true, true}, detects)
	assert.Equal(t, []int{0, 6, 6, 6}, pixels)
}

func TestBackgroundEdges(t *testing.T) {
	camera := new(TestCamera)

	config := defaultMotionParams()
	config.EdgePixels = 2
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	edgeValue := uint16(999)
	backgroundValue := 3000
	g := newFrameGen(detector, camera)
	frame := g.setupFrame(backgroundValue)
	SetFrameEdge(frame, config.EdgePixels, edgeValue)

	detector.updateBackground(frame)
	for y := range frame.Pix {
		for x := 0; x < len(frame.Pix[y]); x++ {
			assert.Equal(t, detector.background.Pix[y][x], uint16(backgroundValue))
		}
	}
	edgeValue = 100
	backgroundValue = 2000
	frame = g.setupFrame(backgroundValue)
	SetFrameEdge(frame, config.EdgePixels, edgeValue)
	detector.updateBackground(frame)

	for y := range frame.Pix {
		for x := 0; x < len(frame.Pix[y]); x++ {
			assert.Equal(t, detector.background.Pix[y][x], uint16(backgroundValue))
		}
	}
}

func SetFrameEdge(frame *cptvframe.Frame, edges int, edgeValue uint16) {
	for y := range frame.Pix {
		cols := edges
		if y < edges || y >= len(frame.Pix)-edges {
			cols = len(frame.Pix[y])
		}
		for x := 0; x < cols; x++ {
			frame.Pix[y][x] = edgeValue
		}
	}
}

func TestCanChangeEdgePixelsValue(t *testing.T) {
	camera := new(TestCamera)

	config := defaultMotionParams()
	config.EdgePixels = 0
	config.WarmerOnly = true
	config.CountThresh = 4
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	detects, pixels := newFrameGen(detector, camera).MovementInColumn(0, 4)
	assert.Equal(t, []bool{false, false, true, true}, detects)
	assert.Equal(t, []int{0, 0, 6, 6}, pixels)
}

func TestSomethingMovingDuringFFC(t *testing.T) {
	camera := new(TestCamera)

	config := defaultMotionParams()
	config.UseOneDiffOnly = true
	config.CountThresh = 4
	detector := NewMotionDetector(config, defaultPreviewFrames(), camera)

	gen := newFrameGen(detector, camera)

	// Fill frame loop.
	detects, pixels := gen.NoMovement(6)
	assertAllFalse(t, detects)
	assertAllZero(t, pixels)

	// Trigger FFC.
	gen.FFC()

	// Because of the FFC, motion should not be reported for the next
	// 10s, even though something warm is moving.
	detects, pixels = gen.Movement(10 * lepton3.FramesHz)
	assertAllFalse(t, detects)
	assertAllZero(t, pixels)

	// Motion is reported again after the FFC period.
	gen.NoMovement(3)
	detects, pixels = gen.Movement(5)
	assert.Equal(t, []bool{false, true, true, true, true}, detects)
	assert.Equal(t, []int{0, 9, 9, 9, 18}, pixels)
}

func defaultMotionParams() config.ThermalMotion {
	return config.ThermalMotion{
		TempThresh:      3000,
		DeltaThresh:     30,
		CountThresh:     8,
		FrameCompareGap: 3,
		WarmerOnly:      false,
		EdgePixels:      1,
	}
}

func defaultPreviewFrames() int {
	return 3 * 9
}

const frameInterval = time.Second / 9

func newFrameGen(detector *motionDetector, camera cptvframe.CameraSpec) *frameGen {
	return &frameGen{
		detector:    detector,
		now:         time.Minute,
		lastFFCTime: time.Second,
		camera:      camera,
	}
}

type frameGen struct {
	detector    *motionDetector
	now         time.Duration
	lastFFCTime time.Duration
	camera      cptvframe.CameraSpec
}

func (g *frameGen) FFC() {
	g.lastFFCTime = g.now
}

func (g *frameGen) NoMovement(frames int) ([]bool, []int) {
	results := make([]bool, frames)
	pixels := make([]int, frames)

	for i := range results {
		frame := g.makeSpot(3300, 0, 0)
		results[i], pixels[i] = g.detector.pixelsChanged(frame, false)
	}
	return results, pixels
}

func (g *frameGen) Movement(frames int) ([]bool, []int) {
	results := make([]bool, frames)
	pixels := make([]int, frames)

	for i := range results {
		frame := g.makeSpot(3300, 10+i, i*100)
		results[i], pixels[i] = g.detector.pixelsChanged(frame, false)
	}
	return results, pixels
}

func (g *frameGen) MovementInColumn(col, frames int) ([]bool, []int) {
	results := make([]bool, frames)
	pixels := make([]int, frames)

	for i := range results {
		log.Println(i)
		frame := g.makeColSpot(3300, 10+5*(i+1), col, (i+1)*100)
		results[i], pixels[i] = g.detector.pixelsChanged(frame, false)
	}
	return results, pixels
}

func (g *frameGen) MovementInRow(row, frames int) ([]bool, []int) {
	results := make([]bool, frames)
	pixels := make([]int, frames)

	for i := range results {
		frame := g.makeRowSpot(3300, row, 10+5*(i+1), (i+1)*100)
		results[i], pixels[i] = g.detector.pixelsChanged(frame, false)
	}
	return results, pixels
}

func (g *frameGen) setupFrame(background int) *cptvframe.Frame {
	frame := cptvframe.NewFrame(g.camera)
	frame.Status.TimeOn = g.now
	frame.Status.LastFFCTime = g.lastFFCTime
	g.now += frameInterval

	// Set background
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			frame.Pix[y][x] = uint16(background)
		}
	}

	return frame
}

func (g *frameGen) makeSpot(background, warmPosition, warmTempOffset int) *cptvframe.Frame {

	frame := g.setupFrame(background)

	// Overlay a warm spot
	warmTemp := uint16(background + warmTempOffset)
	for y := warmPosition; y <= warmPosition+2; y++ {
		for x := warmPosition; x <= warmPosition+2; x++ {
			fmt.Println(y, x, warmTemp)
			frame.Pix[y][x] = warmTemp
		}
	}
	return frame
}

func (g *frameGen) makeColSpot(background, startRow, col, warmTempOffset int) *cptvframe.Frame {

	frame := g.setupFrame(background)

	// Overlay a some of column
	warmTemp := uint16(background + warmTempOffset)
	for y := startRow; y <= startRow+10; y++ {
		frame.Pix[y][col] = warmTemp
	}
	return frame
}

func (g *frameGen) makeRowSpot(background, row, startCol, warmTempOffset int) *cptvframe.Frame {

	frame := g.setupFrame(background)

	// Overlay a some of column
	warmTemp := uint16(background + warmTempOffset)
	for x := startCol; x <= startCol+10; x++ {
		frame.Pix[row][x] = warmTemp
	}
	return frame
}

func assertAllFalse(t *testing.T, b []bool) {
	for _, v := range b {
		if !assert.Equal(t, false, v) {
			return
		}
	}
}

func assertAllZero(t *testing.T, n []int) {
	for _, v := range n {
		if !assert.Equal(t, 0, v) {
			return
		}
	}
}
