// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
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

func TestAnimalRecordings(t *testing.T) {
	config := OldDefaultConfig()

	results := NewCPTVPlaybackTester(config).TestAllCPTVFiles(GetBaseDir() + "/motiontest/animals")

	expectedResults := []string{
		"20180814-153527.cptv Detected: (1:15)(29:66)    Recorded: (2:24)(30:75)    Motion frames: 51/107",
		"20180814-153539.cptv Detected: (1:101)(115:163)(175: Recorded: (2:110)(116:172)(176: Motion frames: 159/190",
		"20180814-182224.cptv Detected: (1:14)(37:37)    Recorded: (2:23)(38:46)    Motion frames: 17/94",
		"cat.cptv             Detected: (25:40)          Recorded: (26:49)          Motion frames: 16/133",
		"rat.cptv             Detected: (1:6)            Recorded: (2:15)           Motion frames: 7/99",
		"rat02.cptv           Detected: (2:21)(57:90)    Recorded: (3:30)(58:99)    Motion frames: 44/107",
		"recalc.cptv          Detected: (1:279)(290:452)(472:479) Recorded: (2:288)(291:461)(473:488) Motion frames: 445/540",
	}

	assert.Equal(t, expectedResults, results)
}

func TestNoiseRecordings(t *testing.T) {
	config := OldDefaultConfig()

	results := NewCPTVPlaybackTester(config).TestAllCPTVFiles(GetBaseDir() + "/motiontest/noise")
	expectedResults := []string{
		"noise_01.cptv        Detected: None             Recorded: None             Motion frames: 0/177",
		"noise_02.cptv        Detected: None             Recorded: None             Motion frames: 0/99",
		"noise_03.cptv        Detected: None             Recorded: None             Motion frames: 1/117",
		"noise_05.cptv        Detected: (34:41)(52:71)(91:93) Recorded: (35:50)(53:80)(92:102) Motion frames: 23/119",
		"skyline.cptv         Detected: None             Recorded: None             Motion frames: 1/91",
	}

	assert.Equal(t, expectedResults, results)
}

func GetBaseDir() string {
	_, file, _, _ := runtime.Caller(0)

	dir, _ := filepath.Abs(filepath.Dir(file))

	return dir
}
