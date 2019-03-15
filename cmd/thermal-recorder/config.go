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
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/TheCacophonyProject/thermal-recorder/throttle"
)

type Config struct {
	DeviceName   string
	FrameInput   string `yaml:"frame-input"`
	OutputDir    string `yaml:"output-dir"`
	MinDiskSpace uint64 `yaml:"min-disk-space"`
	Recorder     recorder.RecorderConfig
	Motion       motion.MotionConfig
	Turret       TurretConfig
	Throttler    throttle.ThrottlerConfig
	Latitude     float32
	Longitude    float32
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

// locationConfig is a struct to store our location values in.
type locationConfig struct {
	Latitude  float32 `yaml:"latitude"`
	Longitude float32 `yaml:"longitude"`
}

func (conf *Config) Validate() error {
	if err := conf.Recorder.Validate(); err != nil {
		return err
	}

	if err := conf.Motion.Validate(); err != nil {
		return err
	}
	return nil
}

var defaultUploaderConfig = uploaderConfig{
	DeviceName: "",
}

var defaultConfig = Config{
	FrameInput:   "/var/run/lepton-frames",
	OutputDir:    "/var/spool/cptv",
	MinDiskSpace: 200,
	Recorder:     recorder.DefaultRecorderConfig(),
	Motion:       motion.DefaultMotionConfig(),
	Throttler:    throttle.DefaultThrottlerConfig(),
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

// ParseLocationFile retrieves values from the location data file.
func parseLocationFile(filepath string) (*locationConfig, error) {

	// Create a default location config
	location := &locationConfig{}

	inBuf, err := ioutil.ReadFile(filepath)
	if os.IsNotExist(err) {
		return location, nil
	} else if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(inBuf, location); err != nil {
		return nil, err
	}
	return location, nil
}

func ParseConfigFiles(recorderFilename, uploaderFilename, locationFileName string) (*Config, error) {
	buf, err := ioutil.ReadFile(recorderFilename)
	if err != nil {
		return nil, err
	}
	uploaderBuf, err := ioutil.ReadFile(uploaderFilename)
	if err != nil {
		return nil, err
	}
	return ParseConfig(buf, uploaderBuf, locationFileName)
}

func ParseConfig(buf, uploaderBuf []byte, locationFileName string) (*Config, error) {
	conf := defaultConfig
	if err := yaml.Unmarshal(buf, &conf); err != nil {
		return nil, err
	}
	uploaderConf := defaultUploaderConfig
	if err := yaml.Unmarshal(uploaderBuf, &uploaderConf); err != nil {
		return nil, err
	}
	location, err := parseLocationFile(locationFileName)
	if err != nil {
		return nil, err
	}

	conf.Latitude = location.Latitude
	conf.Longitude = location.Longitude
	conf.DeviceName = uploaderConf.DeviceName

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return &conf, nil
}
