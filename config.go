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
	SPISpeed  int64  `yaml:"spi-speed"`
	PowerPin  string `yaml:"power-pin"`
	OutputDir string `yaml:"output-dir"`
	MinSecs   int    `yaml:"min-secs"`
	MaxSecs   int    `yaml:"max-secs"`
	Motion    Motion `yaml:"motion"`
}

type Motion struct {
	DeltaThresh uint16 `yaml:"delta-thresh"`
	CountThresh uint16 `yaml:"count-thresh"`
	TempThresh  uint16 `yaml:"temp-thresh"`
}

func (conf *Config) Validate() error {
	if conf.MaxSecs < conf.MinSecs {
		return errors.New("max-secs should be larger than min-secs")
	}
	return nil
}

var defaultConfig = Config{
	SPISpeed:  30000000,
	PowerPin:  "GPIO23",
	OutputDir: "/var/spool/cptv",
	MinSecs:   10,
	MaxSecs:   600,
	Motion: Motion{
		DeltaThresh: 20,
		CountThresh: 10,
		TempThresh:  3200,
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
