// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"io/ioutil"
	"time"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	SPISpeed    int64
	PowerPin    string
	OutputDir   string
	MinSecs     int
	MaxSecs     int
	WindowStart time.Time
	WindowEnd   time.Time
	Motion      MotionConfig
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
}

func (conf *MotionConfig) Validate() error {
	if conf.NonzeroMaxPercent < 1 || conf.NonzeroMaxPercent > 100 {
		return errors.New("nonzero-max-percent should be in range 1 - 100")
	}
	return nil
}

type rawConfig struct {
	SPISpeed    int64        `yaml:"spi-speed"`
	PowerPin    string       `yaml:"power-pin"`
	OutputDir   string       `yaml:"output-dir"`
	MinSecs     int          `yaml:"min-secs"`
	MaxSecs     int          `yaml:"max-secs"`
	WindowStart string       `yaml:"window-start"`
	WindowEnd   string       `yaml:"window-end"`
	Motion      MotionConfig `yaml:"motion"`
}

var defaultConfig = rawConfig{
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
}

func ParseConfigFile(filename string) (*Config, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return ParseConfig(buf)
}

func ParseConfig(buf []byte) (*Config, error) {
	raw := defaultConfig
	if err := yaml.Unmarshal(buf, &raw); err != nil {
		return nil, err
	}

	conf := &Config{
		SPISpeed:  raw.SPISpeed,
		PowerPin:  raw.PowerPin,
		OutputDir: raw.OutputDir,
		MinSecs:   raw.MinSecs,
		MaxSecs:   raw.MaxSecs,
		Motion:    raw.Motion,
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
