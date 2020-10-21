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
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/TheCacophonyProject/window"
)

type TestRecorder struct {
	frameIds         []int
	index            int
	previousFrameIds []int
	CanRecordReturn  error
}

func (tr *TestRecorder) StopRecording() error {
	tr.previousFrameIds = tr.frameIds[:tr.index]
	tr.frameIds = nil
	return nil
}

func (tr *TestRecorder) CheckCanRecord() error { return tr.CanRecordReturn }

func (tr *TestRecorder) StartRecording(background *cptvframe.Frame, tempThresh uint16) error {
	tr.frameIds = make([]int, 200)
	tr.index = 0
	return nil
}

func (tr *TestRecorder) WriteFrame(frame *cptvframe.Frame) error {
	tr.frameIds[tr.index] = int(frame.Pix[0][0])
	tr.index++
	return nil
}

func (tr *TestRecorder) GetRecordedFramesIds() []int {
	return tr.previousFrameIds
}

func (tr *TestRecorder) SetCheckError(err error) {
	tr.CanRecordReturn = err
}

func (tr *TestRecorder) IsRecording() bool {
	return tr.frameIds != nil
}
func LocationTestConfig() *config.Location {
	config := config.Location{}
	return &config
}

func RecorderTestConfig() *recorder.RecorderConfig {
	config := new(recorder.RecorderConfig)
	config.MinSecs = 3
	config.MaxSecs = 20
	config.PreviewSecs = 1
	w, err := window.New("20:00", "20:00", 0, 0)
	if err != nil {
		panic(err)
	}
	config.Window = *w
	return config
}

func MotionTestConfig() *config.ThermalMotion {
	config := new(config.ThermalMotion)

	config.TempThresh = 3000
	config.DeltaThresh = 50
	config.CountThresh = 3
	config.FrameCompareGap = 12
	config.Verbose = false
	config.TriggerFrames = 1
	config.UseOneDiffOnly = true
	config.WarmerOnly = true
	return config
}

func FramesFrom(start, end int) []int {
	slice := make([]int, end-start+1)
	for i := 0; i < end-start+1; i++ {
		slice[i] = i + start
	}
	return slice
}

func SetupTest(mConf *config.ThermalMotion, rConf *recorder.RecorderConfig, lConf *config.Location) (*TestRecorder, *TestFrameMaker) {
	recorder := new(TestRecorder)
	camera := new(TestCamera)
	processor := NewMotionProcessor(lepton3.ParseRawFrame, mConf, rConf, lConf, nil, recorder, camera)

	scenarioMaker := MakeTestFrameMaker(processor, camera)
	return recorder, scenarioMaker
}

func TestRecorderNotTriggeredUnlessSeesMovement(t *testing.T) {
	recorder, scenarioMaker := SetupTest(MotionTestConfig(), RecorderTestConfig(), LocationTestConfig())
	scenarioMaker.AddBackgroundFrames(20)
	assert.False(t, recorder.IsRecording())
}

func TestRecorderTriggeredAndHasPreviewAndMinNumberFrames(t *testing.T) {
	recorder, scenarioMaker := SetupTest(MotionTestConfig(), RecorderTestConfig(), LocationTestConfig())
	scenarioMaker.AddBackgroundFrames(11).AddMovingDotFrames(1).AddBackgroundFrames(40)
	assert.Equal(t, FramesFrom(2, 37), recorder.GetRecordedFramesIds())
}

func TestRecorderNotTriggeredUntilTriggerFramesReached(t *testing.T) {
	config := MotionTestConfig()
	config.TriggerFrames = 3

	recorder, scenarioMaker := SetupTest(config, RecorderTestConfig(), LocationTestConfig())

	// not triggered by 2 moving frames in a row
	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(2).AddBackgroundFrames(8)
	assert.False(t, recorder.IsRecording())

	// not triggered by 3 moving frames in a row
	scenarioMaker.AddMovingDotFrames(3).AddBackgroundFrames(40)
	assert.Equal(t, FramesFrom(11, 48), recorder.GetRecordedFramesIds())
}

func TestRecorderNotStartedIfCheckCanRecordReturnsError(t *testing.T) {
	recorder, scenarioMaker := SetupTest(MotionTestConfig(), RecorderTestConfig(), LocationTestConfig())
	recorder.SetCheckError(errors.New("Cannot record or bad things will happen"))

	// record not triggered due to error return above
	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(2).AddBackgroundFrames(5)
	assert.False(t, recorder.IsRecording())
}

func TestCanMakeMultipleRecordings(t *testing.T) {
	recorder, scenarioMaker := SetupTest(MotionTestConfig(), RecorderTestConfig(), LocationTestConfig())

	scenarioMaker.AddBackgroundFrames(11).AddMovingDotFrames(1).AddBackgroundFrames(39)
	assert.Equal(t, FramesFrom(2, 37), recorder.GetRecordedFramesIds())

	scenarioMaker.AddBackgroundFrames(10).AddMovingDotFrames(1).AddBackgroundFrames(39)
	assert.Equal(t, FramesFrom(52, 87), recorder.GetRecordedFramesIds())
}

func TestMultipleRecordingsDontRepeatAnyFrames(t *testing.T) {
	// if the tail of the previous recording comes within the preview time of the next
	// recording then only the unwritten frames are recorded.
	recorder, scenarioMaker := SetupTest(MotionTestConfig(), RecorderTestConfig(), LocationTestConfig())

	scenarioMaker.AddBackgroundFrames(11).AddMovingDotFrames(1).AddBackgroundFrames(29)
	assert.Equal(t, FramesFrom(2, 37), recorder.GetRecordedFramesIds())

	scenarioMaker.AddMovingDotFrames(1).AddBackgroundFrames(39)
	assert.Equal(t, FramesFrom(38, 67), recorder.GetRecordedFramesIds())
}
