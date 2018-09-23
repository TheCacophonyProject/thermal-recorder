package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func CurrentConfig() *Config {
	config := GetDefaultConfigFromFile()

	// Use smaller min secs to detect more clearly when we stop detecting.
	config.MinSecs = 1

	return config
}

func OldDefaultConfig() *Config {
	config := DefaultTestConfig()
	config.MinSecs = 1
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
		"rat.cptv":      "(1:6)",
		"rat02.cptv":    "(1:14)(61:92)",
		"recalc.cptv":   "(1:497)",
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

func TestCptvNoiseRecordings(t *testing.T) {
	config := CurrentConfig()

	actualResults := NewCPTVPlaybackTester(config).TestAllCPTVFiles(GetBaseDir() + "/motiontest/noise")

	expectedResults := map[string]string{
		"noise_01.cptv": "None",
		"noise_02.cptv": "None",
		"noise_03.cptv": "(75:79)",
		"noise_05.cptv": "(19:76)(90:94)",
		"skyline.cptv":  "None",
	}

	CompareDetectedPeriods(t, expectedResults, actualResults)
}

// DoTestResearchAnimalRecordings - change this to test to run though different scenarios of test
// calculations.   It will output the results to /motiontest/results
func DoTestResearchAnimalRecordings(t *testing.T) {
	testname := "cut off - animals2"
	searchDir := GetBaseDir() + "/motiontest/animals"

	f, err := os.Create(GetBaseDir() + "/motiontest/results/" + testname)
	if err != nil {
		return
	}
	defer f.Close()
	writer := bufio.NewWriter(f)

	config := CurrentConfig()
	config.Motion.TempThresh = 2900
	ExperimentAndWriteResultsToFile(testname+"2900", config, searchDir, writer)

	config.Motion.TempThresh = 2800
	ExperimentAndWriteResultsToFile(testname+"2800", config, searchDir, writer)

	config.Motion.TempThresh = 2700
	ExperimentAndWriteResultsToFile(testname+"2700", config, searchDir, writer)

	config.Motion.TempThresh = 2500
	ExperimentAndWriteResultsToFile(testname+"2500", config, searchDir, writer)

	ExperimentAndWriteResultsToFile("Current config", CurrentConfig(), searchDir, writer)

	ExperimentAndWriteResultsToFile("Old default", OldDefaultConfig(), searchDir, writer)
}

func ExperimentAndWriteResultsToFile(name string, config *Config, dir string, writer *bufio.Writer) {
	fmt.Fprintf(writer, "Results for %s", name)
	fmt.Fprintln(writer)

	results := NewCPTVPlaybackTester(config).TestAllCPTVFiles(dir)

	fmt.Fprintf(writer, "%-10s:  %ds - %ds", "Recording limits", config.MinSecs, config.MaxSecs)
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

	recorder := new(NoWriteRecorder)

	processor := NewMotionProcessor(config, nil, recorder)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, frame := range frames {
			processor.processFrame(frame)
		}
	}
}
