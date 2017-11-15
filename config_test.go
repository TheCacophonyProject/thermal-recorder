package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllDefaults(t *testing.T) {
	conf, err := ParseConfig([]byte(""))
	require.NoError(t, err)
	require.NoError(t, conf.Validate())

	assert.Equal(t, Config{
		SPISpeed:  2500000,
		PowerPin:  "GPIO23",
		OutputDir: "/var/spool/cptv",
		MinSecs:   10,
		MaxSecs:   600,
		Motion: MotionConfig{
			TempThresh:        3000,
			DeltaThresh:       30,
			CountThresh:       5,
			NonzeroMaxPercent: 50,
		},
		LEDs: LEDsConfig{
			Recording: "GPIO20",
			Running:   "GPIO21",
		},
	}, *conf)
}

func TestAllSet(t *testing.T) {
	// All config set at non-default values.
	config := []byte(`
spi-speed: 123
power-pin: "PIN"
output-dir: "/some/where"
min-secs: 2
max-secs: 10
window-start: 17:10
window-end: 07:20
motion:
    temp-thresh: 2000
    delta-thresh: 20
    count-thresh: 1
    nonzero-max-percent: 20
leds:
    recording: "RecordingPIN"
    running: "RunningPIN"
`)

	conf, err := ParseConfig(config)
	require.NoError(t, err)
	require.NoError(t, conf.Validate())

	assert.Equal(t, Config{
		SPISpeed:    123,
		PowerPin:    "PIN",
		OutputDir:   "/some/where",
		MinSecs:     2,
		MaxSecs:     10,
		WindowStart: time.Date(0, 1, 1, 17, 10, 0, 0, time.UTC),
		WindowEnd:   time.Date(0, 1, 1, 07, 20, 0, 0, time.UTC),
		Motion: MotionConfig{
			TempThresh:        2000,
			DeltaThresh:       20,
			CountThresh:       1,
			NonzeroMaxPercent: 20,
		},
		LEDs: LEDsConfig{
			Recording: "RecordingPIN",
			Running:   "RunningPIN",
		},
	}, *conf)
}

func TestInvalidWindowStart(t *testing.T) {
	conf, err := ParseConfig([]byte("window-start: 25:10"))
	assert.Nil(t, conf)
	assert.EqualError(t, err, "invalid window-start")
}

func TestInvalidWindowEnd(t *testing.T) {
	conf, err := ParseConfig([]byte("window-end: 25:10"))
	assert.Nil(t, conf)
	assert.EqualError(t, err, "invalid window-end")
}

func TestWindowEndWithoutStart(t *testing.T) {
	conf, err := ParseConfig([]byte("window-end: 09:10"))
	assert.Nil(t, conf)
	assert.EqualError(t, err, "window-end is set but window-start isn't")
}

func TestWindowStartWithoutEnd(t *testing.T) {
	conf, err := ParseConfig([]byte("window-start: 09:10"))
	assert.Nil(t, conf)
	assert.EqualError(t, err, "window-start is set but window-end isn't")
}
