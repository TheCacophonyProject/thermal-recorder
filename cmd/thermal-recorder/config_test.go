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
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/TheCacophonyProject/thermal-recorder/location"
	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/TheCacophonyProject/thermal-recorder/throttle"
	"github.com/TheCacophonyProject/window"
)

const (
	testDir = "test_data"
)

var testLocationFileName = filepath.Join(GetBaseDir(), testDir, "location.yaml")
var testConfigFile = filepath.Join(GetBaseDir(), testDir, "config.yaml")
var testUploaderFile = filepath.Join(GetBaseDir(), testDir, "uploader-config.yaml")

func getExpectedDefaultConfig() Config {
	return Config{
		DeviceName:   "",
		Location:     location.DefaultLocationConfig(),
		FrameInput:   "/var/run/lepton-frames",
		OutputDir:    "/var/spool/cptv",
		MinDiskSpace: 200,
		Recorder: recorder.RecorderConfig{
			MinSecs:     10,
			MaxSecs:     600,
			PreviewSecs: 3,
		},
		Motion: motion.MotionConfig{
			TempThresh:      2900,
			DeltaThresh:     50,
			CountThresh:     3,
			FrameCompareGap: 45,
			UseOneDiffOnly:  true,
			Verbose:         false,
			TriggerFrames:   2,
			WarmerOnly:      true,
			EdgePixels:      1,
		},
		Throttler: throttle.ThrottlerConfig{
			ApplyThrottling: true,
			ThrottleAfter:   600,
			MinRefill:       10 * time.Minute,
		},
		Turret: TurretConfig{
			Active: false,
			PID:    []float64{0.05, 0, 0},
			ServoX: ServoConfig{
				Active:   false,
				Pin:      "17",
				MaxAng:   160,
				MinAng:   20,
				StartAng: 90,
			},
			ServoY: ServoConfig{
				Active:   false,
				Pin:      "18",
				MaxAng:   160,
				MinAng:   20,
				StartAng: 90,
			},
		},
	}
}

func TestAllDefaults(t *testing.T) {
	conf, err := ParseConfig([]byte(""), []byte(""), []byte(""))
	expected := getExpectedDefaultConfig()
	require.NoError(t, err)
	require.NoError(t, conf.Validate())

	assert.Equal(t, expected, *conf)
}

func TestDefaultsMatchDefaultYamlFile(t *testing.T) {
	configDefaults, err := ParseConfig([]byte(""), []byte(""), []byte(""))
	require.NoError(t, err)

	defaultConfig := GetDefaultConfig()
	var configYAML Config
	yaml.UnmarshalStrict(defaultConfig, &configYAML)
	// ignore errors in turret since they aren't in use atm
	configDefaults.Turret = configYAML.Turret

	configYAML.Location = location.DefaultLocationConfig()
	assert.Equal(t, configDefaults, &configYAML)
}

func getExpectedAllSetConfig() Config {
	return Config{
		DeviceName: "aDeviceName",
		Location: location.LocationConfig{
			Latitude:  -36,
			Longitude: 174,
			Accuracy:  10,
		},
		FrameInput:   "/some/sock",
		OutputDir:    "/some/where",
		MinDiskSpace: 321,
		Recorder: recorder.RecorderConfig{
			MinSecs:     2,
			MaxSecs:     10,
			PreviewSecs: 5,
			WindowStart: *window.NewTimeOfDay("17:10"),
			WindowEnd:   *window.NewTimeOfDay("07:20"),
		},
		Motion: motion.MotionConfig{
			TempThresh:      2000,
			DeltaThresh:     20,
			CountThresh:     1,
			FrameCompareGap: 90,
			UseOneDiffOnly:  false,
			Verbose:         true,
			TriggerFrames:   1,
			WarmerOnly:      false,
			EdgePixels:      3,
		},
		Throttler: throttle.ThrottlerConfig{
			ApplyThrottling: false,
			ThrottleAfter:   650,
			MinRefill:       15 * time.Minute,
		},
		Turret: TurretConfig{
			Active: true,
			PID:    []float64{1, 2, 3},
			ServoX: ServoConfig{
				Active:   false,
				Pin:      "pin",
				MaxAng:   180,
				MinAng:   0,
				StartAng: 30,
			},
			ServoY: ServoConfig{
				Active:   true,
				Pin:      "pin1",
				MaxAng:   190,
				MinAng:   10,
				StartAng: 40,
			},
		},
	}
}

func TestAllSet(t *testing.T) {
	conf, err := ParseConfigFiles(testConfigFile, testUploaderFile, testLocationFileName)
	expected := getExpectedAllSetConfig()
	require.NoError(t, err)
	require.NoError(t, conf.Validate())
	assert.Equal(t, expected, *conf)
}

func GetDefaultConfig() []byte {
	locationBuf, _ := ioutil.ReadFile(GetDefaultConfigFile())
	return locationBuf
}

func GetDefaultConfigFile() string {
	dir := GetBaseDir()
	configFile := strings.Replace(dir, filepath.Join("cmd/thermal-recorder"), filepath.Join("_release/thermal-recorder.yaml"), 1)
	return configFile
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

func getRecorerErrorConfig() Config {
	return Config{
		Recorder: recorder.RecorderConfig{
			MinSecs: 10,
			MaxSecs: 4,
		},
	}
}

func TestRecorderErrorsStopConfigParsing(t *testing.T) {
	conf := getRecorerErrorConfig()
	err := conf.Validate()
	assert.EqualError(t, err, "max-secs should be larger than min-secs")
}
