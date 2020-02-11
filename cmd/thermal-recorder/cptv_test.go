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

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/lepton3"
	"github.com/TheCacophonyProject/window"
	"github.com/stretchr/testify/assert"

	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
)

func CurrentConfig() *Config {
	//GetDefaultConfig()
	w, _ := window.New("12:00", "12:00", 0, 0)
	recorder := recorder.RecorderConfig{
		MaxSecs:     config.DefaultThermalRecorder().MaxSecs,
		MinSecs:     1, // Use smaller min secs to detect more clearly when we stop detecting.
		PreviewSecs: config.DefaultThermalRecorder().PreviewSecs,
		Window:      *w,
	}
	return &Config{
		DeviceName:   "test name",
		FrameInput:   config.DefaultLepton().FrameOutput,
		Location:     config.Location{},
		MinDiskSpace: config.DefaultThermalRecorder().MinDiskSpaceMB,
		Motion:       config.DefaultThermalMotion(),
		OutputDir:    config.DefaultThermalRecorder().OutputDir,
		Recorder:     recorder,
		Throttler:    config.DefaultThermalThrottler(),
	}
}

func OldDefaultConfig() *Config {
	config := new(Config)
	config.Recorder.MinSecs = 1
	config.Recorder.MaxSecs = 20
	config.Recorder.PreviewSecs = 1
	config.Motion.TriggerFrames = 1
	config.Motion.UseOneDiffOnly = false
	config.Motion.FrameCompareGap = 1
	config.Motion.DeltaThresh = 30
	config.Motion.CountThresh = 5
	config.Motion.WarmerOnly = false
	return config
}

func TestCptvAnimalRecordings(t *testing.T) {
	config := CurrentConfig()

	actualResults := NewCPTVPlaybackTester(config).TestAllCPTVFiles(GetBaseDir() + "/motiontest/animals")

	expectedResults := map[string]string{
		"cat.cptv":      "(25:41)",
		"hedgehog.cptv": "(3:32)(45:end)",
		"possum02.cptv": "(1:end)",
		"rat.cptv":      "(1:6)(73:84)",
		"rat02.cptv":    "(1:23)(57:90)",
	}

	CompareDetectedPeriods(t, expectedResults, actualResults)
}

func CompareDetectedPeriods(t *testing.T, expectedResults map[string]string, actual map[string]*EventLoggingRecordingListener) {
	errors := 0

	for key, expected := range expectedResults {
		if actual, ok := actual[key]; !ok {
			log.Printf("Missing results for file  %s", key)
			errors++
		} else if expected != actual.motionDetectedFrames {
			log.Printf("Expected results for %-16s: %s", key, expected)
			log.Printf("Actual results for   %-16s: %s", key, actual.motionDetectedFrames)
			errors++
		}
	}

	for key, result := range actual {
		if _, ok := expectedResults[key]; ok == false {
			log.Printf("Extra file %s has results %s", key, result.motionDetectedFrames)
			errors++
		}
	}

	if errors > 0 {
		assert.Fail(t, fmt.Sprintf("There were %d errors.", errors))
	}
}

func GetBaseDir() string {
	_, file, _, ok := runtime.Caller(0)

	if !ok {
		panic(fmt.Errorf("Could not find the base dir where sample files are"))
	}

	dir, err := filepath.Abs(filepath.Dir(file))

	if err != nil {
		panic(err)
	}

	return dir
}

func TestCptvNoiseRecordings(t *testing.T) {
	config := CurrentConfig()

	actualResults := NewCPTVPlaybackTester(config).TestAllCPTVFiles(GetBaseDir() + "/motiontest/noise")

	expectedResults := map[string]string{
		"noise_01.cptv": "None",
		"noise_02.cptv": "(1:53)",
		"noise_03.cptv": "(75:79)",
		"noise_05.cptv": "(60:62)(91:94)",
		"skyline.cptv":  "None",
	}

	CompareDetectedPeriods(t, expectedResults, actualResults)
}

func TestCptvFunnyEdgeNoise(t *testing.T) {
	config := CurrentConfig()
	config.Motion.DeltaThresh = 40

	actualResults := NewCPTVPlaybackTester(config).TestAllCPTVFiles(GetBaseDir() + "/motiontest/edge")

	expectedResults := map[string]string{
		"20181123-022114.cptv": "None",
	}

	CompareDetectedPeriods(t, expectedResults, actualResults)
}

// DoTestResearchAnimalRecordings - change this to test to run though different scenarios of test
// calculations.   It will output the results to /motiontest/results
func DoTestResearchAnimalRecordings(t *testing.T) {
	testname := "lines "
	searchDir := GetBaseDir() + "/motiontest/adhoc/lines"

	f, err := os.Create(GetBaseDir() + "/motiontest/results/" + testname)
	if err != nil {
		return
	}
	defer f.Close()
	writer := bufio.NewWriter(f)

	config := CurrentConfig()
	config.Motion.EdgePixels = 0
	config.Motion.DeltaThresh = 40
	config.Motion.Verbose = true

	ExperimentAndWriteResultsToFile(testname+"0", config, searchDir, writer)

	config.Motion.EdgePixels = 1
	ExperimentAndWriteResultsToFile(testname+"1", config, searchDir, writer)

	t.Fail()
}

func ExperimentAndWriteResultsToFile(name string, config *Config, dir string, writer *bufio.Writer) {
	fmt.Fprintf(writer, "Results for %s", name)
	fmt.Fprintln(writer)

	results := NewCPTVPlaybackTester(config).TestAllCPTVFiles(dir)

	fmt.Fprintf(writer, "%-10s:  %ds - %ds", "Recording limits", config.Recorder.MinSecs, config.Recorder.MaxSecs)
	fmt.Fprintln(writer)
	fmt.Fprintf(writer, "Motion: %+v", config.Motion)
	fmt.Fprintln(writer)

	for key, result := range results {
		fmt.Fprintf(writer, "%-20s Detected: %-16s Recorded: %-16s Motion frames: %d/%d",
			key, result.motionDetectedFrames,
			result.recordedFrames,
			result.motionDetectedCount,
			result.frameCount)
		fmt.Fprintln(writer)
	}

	fmt.Fprintln(writer)
	fmt.Fprintln(writer)

	writer.Flush()
}

func BenchmarkMotionDetection(b *testing.B) {
	config := CurrentConfig()

	tester := NewCPTVPlaybackTester(config)
	frames := tester.LoadAllCptvFrames(GetBaseDir() + "/motiontest/animals/recalc.cptv")

	recorder := new(recorder.NoWriteRecorder)

	processor := motion.NewMotionProcessor(lepton3.ParseRawFrame, &config.Motion, &config.Recorder, &config.Location, nil, recorder, new(TestCamera))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, frame := range frames {
			processor.ProcessFrame(frame)
		}
	}
}
