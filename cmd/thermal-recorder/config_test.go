package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"

	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/TheCacophonyProject/thermal-recorder/throttle"
	"github.com/TheCacophonyProject/window"
)

func TestAllDefaults(t *testing.T) {
	conf, err := ParseConfig([]byte(""), []byte(""))
	require.NoError(t, err)
	require.NoError(t, conf.Validate())

	assert.Equal(t, Config{
		DeviceName:   "",
		FrameInput:   "/var/run/lepton-frames",
		OutputDir:    "/var/spool/cptv",
		MinDiskSpace: 200,
		Recorder: recorder.RecorderConfig{
			MinSecs:     10,
			MaxSecs:     600,
			PreviewSecs: 3,
		},
		Motion: motion.MotionConfig{
			TempThresh:        2900,
			DeltaThresh:       50,
			CountThresh:       3,
			NonzeroMaxPercent: 50,
			FrameCompareGap:   45,
			UseOneDiffOnly:    true,
			Verbose:           false,
			TriggerFrames:     2,
			WarmerOnly:        true,
		},
		Throttler: throttle.ThrottlerConfig{
			ApplyThrottling: true,
			ThrottleAfter:   600,
			SparseAfter:     3600,
			SparseLength:    30,
			RefillRate:      1,
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
	}, *conf)
}

func TestAllProgramDefaultsMatchDefaultYamlFile(t *testing.T) {
	configDefaults, err := ParseConfig([]byte(""), []byte(""))
	require.NoError(t, err)

	defaultConfig := GetDefaultConfig()
	var configYAML Config
	yaml.UnmarshalStrict(defaultConfig, &configYAML)

	// ignore errors in turret since they aren't in use atm
	configDefaults.Turret = configYAML.Turret

	assert.Equal(t, configDefaults, &configYAML)
}

func TestAllSet(t *testing.T) {
	// All config set at non-default values.
	config := []byte(`
frame-input: "/some/sock"
output-dir: "/some/where"
min-disk-space: 321
recorder:
    min-secs: 2
    max-secs: 10
    preview-secs: 5
    window-start: 17:10
    window-end: 07:20
motion:
    temp-thresh: 2000
    delta-thresh: 20
    count-thresh: 1
    nonzero-max-percent: 20
    frame-compare-gap: 90
    one-diff-only: false
    trigger-frames: 1
    verbose: true
    warmer-only: false
throttler:
    apply-throttling: false
    throttle-after-secs: 650
    sparse-after-secs: 6500
    sparse-length-secs: 300
    refill-rate: 0.2
leds:
    recording: "RecordingPIN"
    running: "RunningPIN"
turret:
    active: true
    pid:
      - 1
      - 2
      - 3
    servo-x:
      active: false
      pin: "pin"
      min-ang: 0
      max-ang: 180
      start-ang: 30
    servo-y:
      active: true
      pin: "pin1"
      min-ang: 10
      max-ang: 190
      start-ang: 40
`)

	uploaderConfig := []byte(`
device-name: "aDeviceName"
`)

	conf, err := ParseConfig(config, uploaderConfig)
	require.NoError(t, err)
	require.NoError(t, conf.Validate())

	assert.Equal(t, Config{
		DeviceName:   "aDeviceName",
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
			TempThresh:        2000,
			DeltaThresh:       20,
			CountThresh:       1,
			NonzeroMaxPercent: 20,
			FrameCompareGap:   90,
			UseOneDiffOnly:    false,
			Verbose:           true,
			TriggerFrames:     1,
			WarmerOnly:        false,
		},
		Throttler: throttle.ThrottlerConfig{
			ApplyThrottling: false,
			ThrottleAfter:   650,
			SparseAfter:     6500,
			SparseLength:    300,
			RefillRate:      0.2,
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
	}, *conf)
}

func GetDefaultConfig() []byte {
	dir := GetBaseDir()
	config_file := strings.Replace(dir, "cmd/thermal-recorder", "_release/thermal-recorder.yaml", 1)
	buf, err := ioutil.ReadFile(config_file)
	if err != nil {
		panic(err)
	}
	return buf
}

func GetDefaultConfigFromFile() *Config {
	config, err := ParseConfig(GetDefaultConfig(), []byte(""))
	if err != nil {
		panic(err)
	}
	return config
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
func TestRecorderErrorsStopConfigParsing(t *testing.T) {
	configStr := []byte(`
recorder:
  min-secs: 10
  max-secs: 4
`)
	conf, err := ParseConfig(configStr, []byte(""))
	assert.Nil(t, conf)
	assert.EqualError(t, err, "max-secs should be larger than min-secs")
}
