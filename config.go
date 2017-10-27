// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"errors"
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	SPISpeed  int64        `yaml:"spi-speed"`
	PowerPin  string       `yaml:"power-pin"`
	OutputDir string       `yaml:"output-dir"`
	MinSecs   int          `yaml:"min-secs"`
	MaxSecs   int          `yaml:"max-secs"`
	Motion    MotionConfig `yaml:"motion"`
	LEDs      LEDsConfig   `yaml:"leds"`
}

type MotionConfig struct {
	TempThresh        uint16 `yaml:"temp-thresh"`
	DeltaThresh       uint16 `yaml:"delta-thresh"`
	CountThresh       int    `yaml:"count-thresh"`
	NonzeroMaxPercent int    `yaml:"nonzero-max-percent"`
}

type LEDsConfig struct {
	Recording string `yaml:"recording"`
	Power     string `yaml:"power"`
}

func (conf *Config) Validate() error {
	if conf.MaxSecs < conf.MinSecs {
		return errors.New("max-secs should be larger than min-secs")
	}
	if err := conf.Motion.Validate(); err != nil {
		return err
	}
	return nil
}

func (conf *MotionConfig) Validate() error {
	if conf.NonzeroMaxPercent < 1 || conf.NonzeroMaxPercent > 100 {
		return errors.New("nonzero-max-percent should be in range 1 - 100")
	}
	return nil
}

var defaultConfig = Config{
	SPISpeed:  25000000,
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
		Power:     "GPIO21",
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
	conf := defaultConfig
	if err := yaml.Unmarshal(buf, &conf); err != nil {
		return nil, err
	}
	if err := conf.Validate(); err != nil {
		return nil, err
	}
	return &conf, nil
}
