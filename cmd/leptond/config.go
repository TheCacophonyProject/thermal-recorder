// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"io/ioutil"

	yaml "gopkg.in/yaml.v2"
)

type Config struct {
	SPISpeed    int64  `yaml:"spi-speed"`
	PowerPin    string `yaml:"power-pin"`
	FrameOutput string `yaml:"frame-output"`
}

var defaultConfig = Config{
	SPISpeed:    2000000,
	PowerPin:    "GPIO23",
	FrameOutput: "/var/run/lepton-frames",
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
	return &conf, nil
}
