package main

import (
	"errors"
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	DeviceName   string
	FrameInput   string
	OutputDir    string
	MinSecs      int
	MaxSecs      int
	PreviewSecs  int
	WindowStart  time.Time
	WindowEnd    time.Time
	MinDiskSpace uint64
	Motion       MotionConfig
	Turret       TurretConfig
}

type ServoConfig struct {
	Active   bool    `yaml:"active"`
	MinAng   float64 `yaml:"min-ang"`
	MaxAng   float64 `yaml:"max-ang"`
	StartAng float64 `yaml:"start-ang"`
	Pin      string  `yaml:"pin"`
}

type TurretConfig struct {
	Active bool        `yaml:"active"`
	PID    []float64   `yaml:"pid"`
	ServoX ServoConfig `yaml:"servo-x"`
	ServoY ServoConfig `yaml:"servo-y"`
}

type uploaderConfig struct {
	DeviceName string `yaml:"device-name"`
}

func (conf *Config) Validate() error {
	if conf.MaxSecs < conf.MinSecs {
		return errors.New("max-secs should be larger than min-secs")
	}
	if conf.WindowStart.IsZero() && !conf.WindowEnd.IsZero() {
		return errors.New("window-end is set but window-start isn't")
	}
	if !conf.WindowStart.IsZero() && conf.WindowEnd.IsZero() {
		return errors.New("window-start is set but window-end isn't")
	}
	if err := conf.Motion.Validate(); err != nil {
		return err
	}
	return nil
}

type MotionConfig struct {
	TempThresh        uint16 `yaml:"temp-thresh"`
	DeltaThresh       uint16 `yaml:"delta-thresh"`
	CountThresh       int    `yaml:"count-thresh"`
	NonzeroMaxPercent int    `yaml:"nonzero-max-percent"`
	FrameCompareGap   int    `yaml:"frame-compare-gap"`
	UseOneDiffOnly    bool   `yaml:"one-diff-only"`
	TriggerFrames     int    `yaml:"trigger-frames"`
	WarmerOnly        bool   `yaml:"warmer-only"`
	Verbose           bool   `yaml:"verbose"`
}

func (conf *MotionConfig) Validate() error {
	if conf.NonzeroMaxPercent < 1 || conf.NonzeroMaxPercent > 100 {
		return errors.New("nonzero-max-percent should be in range 1 - 100")
	}
	return nil
}

type rawConfig struct {
	FrameInput   string       `yaml:"frame-input"`
	OutputDir    string       `yaml:"output-dir"`
	MinSecs      int          `yaml:"min-secs"`
	MaxSecs      int          `yaml:"max-secs"`
	PreviewSecs  int          `yaml:"preview-secs"`
	WindowStart  string       `yaml:"window-start"`
	WindowEnd    string       `yaml:"window-end"`
	MinDiskSpace uint64       `yaml:"min-disk-space"`
	Motion       MotionConfig `yaml:"motion"`
	Turret       TurretConfig `yaml:"turret"`
}

var defaultUploaderConfig = uploaderConfig{
	DeviceName: "",
}

var defaultConfig = rawConfig{
	FrameInput:   "/var/run/lepton-frames",
	OutputDir:    "/var/spool/cptv",
	MinSecs:      10,
	MaxSecs:      600,
	PreviewSecs:  3,
	MinDiskSpace: 200,
	Motion: MotionConfig{
		TempThresh:        2900,
		DeltaThresh:       50,
		CountThresh:       3,
		NonzeroMaxPercent: 50,
		FrameCompareGap:   45,
		Verbose:           false,
		TriggerFrames:     2,
		UseOneDiffOnly:    true,
		WarmerOnly:        true,
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

func ParseConfigFiles(recorderFilename, uploaderFilename string) (*Config, error) {
	buf, err := ioutil.ReadFile(recorderFilename)
	if err != nil {
		return nil, err
	}
	uploaderBuf, err := ioutil.ReadFile(uploaderFilename)
	if err != nil {
		return nil, err
	}
	return ParseConfig(buf, uploaderBuf)
}

func ParseConfig(buf, uploaderBuf []byte) (*Config, error) {
	raw := defaultConfig
	if err := yaml.Unmarshal(buf, &raw); err != nil {
		return nil, err
	}
	uploaderConf := defaultUploaderConfig
	if err := yaml.Unmarshal(uploaderBuf, &uploaderConf); err != nil {
		return nil, err
	}

	conf := &Config{
		DeviceName:   uploaderConf.DeviceName,
		FrameInput:   raw.FrameInput,
		OutputDir:    raw.OutputDir,
		MinSecs:      raw.MinSecs,
		MaxSecs:      raw.MaxSecs,
		PreviewSecs:  raw.PreviewSecs,
		MinDiskSpace: raw.MinDiskSpace,
		Motion:       raw.Motion,
		Turret:       raw.Turret,
	}

	const timeOnly = "15:04"
	if raw.WindowStart != "" {
		t, err := time.Parse(timeOnly, raw.WindowStart)
		if err != nil {
			return nil, errors.New("invalid window-start")
		}
		conf.WindowStart = t
	}
	if raw.WindowEnd != "" {
		t, err := time.Parse(timeOnly, raw.WindowEnd)
		if err != nil {
			return nil, errors.New("invalid window-end")
		}
		conf.WindowEnd = t
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return conf, nil
}
