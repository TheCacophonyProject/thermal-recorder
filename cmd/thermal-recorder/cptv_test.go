// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CurrentConfig() *Config {
	dir := GetBaseDir()
	config_file := strings.Replace(dir, "cmd/thermal-recorder", "thermal-recorder-TEMPLATE.yaml", 1)
	uploader_config_file := dir + "/motiontest/thermal-uploader-test.yaml"
	config, _ := ParseConfigFiles(config_file, uploader_config_file)
	// Use smaller min secs to detect more clearly when we stop detecting.
	config.MinSecs = 1
	logConfig(config)

	return config
}

func OldDefaultConfig() *Config {
	config := DefaultTestConfig()
	config.MinSecs = 1
	config.Motion.TriggerFrames = 1
	config.Motion.UseOneFrameOnly = false
	config.Motion.FrameCompareGap = 1
	config.Motion.DeltaThresh = 30
	config.Motion.CountThresh = 5
	config.Motion.WarmerOnly = false
	return config
}

func TestCptvAnimalRecordings(t *testing.T) {
	config := CurrentConfig()

	results := NewCPTVPlaybackTester(config).TestAllCPTVFiles(GetBaseDir() + "/motiontest/animals")

	expectedResults := []string{
		"cat.cptv             Detected: (25:41)          Recorded: (26:50)          Motion frames: 18/133",
		"hedgehog.cptv        Detected: (3:32)(45:       Recorded: (4:41)(46:       Motion frames: 55/100",
		"possum02.cptv        Detected: (1:              Recorded: (2:              Motion frames: 92/94",
		"rat.cptv             Detected: (1:6)            Recorded: (2:15)           Motion frames: 7/99",
		"rat02.cptv           Detected: (1:14)(61:92)    Recorded: (2:23)(62:101)   Motion frames: 39/107",
		"recalc.cptv          Detected: (1:497)          Recorded: (2:506)          Motion frames: 490/540",
	}

	assert.Equal(t, expectedResults, results)
}

func TestCptvNoiseRecordings(t *testing.T) {
	config := CurrentConfig()

	results := NewCPTVPlaybackTester(config).TestAllCPTVFiles(GetBaseDir() + "/motiontest/noise")
	expectedResults := []string{
		"noise_01.cptv        Detected: None             Recorded: None             Motion frames: 1/177",
		"noise_02.cptv        Detected: None             Recorded: None             Motion frames: 1/99",
		"noise_03.cptv        Detected: (75:79)          Recorded: (76:88)          Motion frames: 9/117",
		"noise_05.cptv        Detected: (19:76)(90:94)   Recorded: (20:85)(91:103)  Motion frames: 43/119",
		"skyline.cptv         Detected: None             Recorded: None             Motion frames: 1/91",
	}

	assert.Equal(t, expectedResults, results)
}

// DoTestResearchAnimalRecordings - change this to test to run though different scenarios of test
// calculations.   It will output the results to /motiontest/results
func DoTestResearchAnimalRecordings(t *testing.T) {
	testname := "cut off - noise"
	searchDir := GetBaseDir() + "/motiontest/noise"

	f, err := os.Create(GetBaseDir() + "/motiontest/results/" + testname)
	if err != nil {
		return
	}
	defer f.Close()
	writer := bufio.NewWriter(f)

	config := CurrentConfig()
	config.Motion.TempThresh = 2900
	ExperiementAndWriteResultsToFile(testname+"2900", config, searchDir, writer)

	config.Motion.TempThresh = 2800
	ExperiementAndWriteResultsToFile(testname+"2800", config, searchDir, writer)

	config.Motion.TempThresh = 2700
	ExperiementAndWriteResultsToFile(testname+"2700", config, searchDir, writer)

	config.Motion.TempThresh = 2500
	ExperiementAndWriteResultsToFile(testname+"2500", config, searchDir, writer)

	ExperiementAndWriteResultsToFile("Current config", CurrentConfig(), searchDir, writer)

	ExperiementAndWriteResultsToFile("Old default", OldDefaultConfig(), searchDir, writer)
}

func ExperiementAndWriteResultsToFile(name string, config *Config, dir string, writer *bufio.Writer) {
	fmt.Fprintf(writer, "Results for %s", name)
	fmt.Fprintln(writer)

	results := NewCPTVPlaybackTester(config).TestAllCPTVFiles(dir)

	fmt.Fprintf(writer, "%-10s:  %ds - %ds", "Recording limits", config.MinSecs, config.MaxSecs)
	fmt.Fprintln(writer, "")
	fmt.Fprintln(writer, config.Motion)

	for ii := range results {
		fmt.Fprintln(writer, results[ii])
	}
	fmt.Fprintln(writer)
	fmt.Fprintln(writer)

	writer.Flush()
}

func GetBaseDir() string {
	_, file, _, _ := runtime.Caller(0)

	dir, _ := filepath.Abs(filepath.Dir(file))

	return dir
}

func BenchmarkMotionDetection(b *testing.B) {
	config := CurrentConfig()

	tester := NewCPTVPlaybackTester(config)
	frames := tester.LoadAllCptvFrames(GetBaseDir() + "/motiontest/animals/recalc.cptv")

	listener := new(HardwareListener) // currently doesn't do anything
	recorder := new(NoWriteRecorder)

	processor := NewMotionProcessor(config, listener, recorder)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for jj := range frames {
			processor.processFrame(frames[jj])
		}
	}
}
