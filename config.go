// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package main

import (
	"io"
	"os"

	"github.com/BurntSushi/toml"
)

type Config struct {
	SPISpeed  int64  `toml:"spi-speed"`
	PowerPin  string `toml:"power-pin"`
	OutputDir string `toml:"output-dir"`
	MinSecs   int    `toml:"min-secs"`
	MaxSecs   int    `toml:"max-secs"`
	Motion    Motion `toml:"motion"`
}

type Motion struct {
	DeltaThresh uint16 `toml:"delta-thresh"`
	CountThresh uint16 `toml:"count-thresh"`
	TempThresh  uint16 `toml:"temp-thresh"`
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

func ConfigFromFile(filename string) (*Config, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ConfigFromReader(f)
}

func ConfigFromReader(r io.Reader) (*Config, error) {
	conf := defaultConfig
	if _, err := toml.DecodeReader(r, &conf); err != nil {
		return nil, err
	}
	return &conf, nil
}
