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

	"github.com/TheCacophonyProject/thermal-recorder/throttle"
	yaml "gopkg.in/yaml.v2"

	"github.com/TheCacophonyProject/thermal-recorder/location"
	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
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
	Location     location.LocationConfig
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
	if err := conf.Recorder.Validate(); err != nil {
		return err
	}

	if err := conf.Motion.Validate(); err != nil {
		return err
	}

	if err := conf.Location.Validate(); err != nil {
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
	Location:     location.DefaultLocationConfig(),
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

func ParseConfigFiles(recorderFilename, uploaderFilename, locationFileName string) (*Config, error) {
	buf, err := ioutil.ReadFile(recorderFilename)
	if err != nil {
		return nil, err
	}

	uploaderBuf, err := ioutil.ReadFile(uploaderFilename)
	if err != nil {
		return nil, err
	}

	locationBuf, err := ioutil.ReadFile(locationFileName)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return ParseConfig(buf, uploaderBuf, locationBuf)
}

func ParseConfig(buf, uploaderBuf, locationBuf []byte) (*Config, error) {

	conf := defaultConfig
	if err := yaml.Unmarshal(buf, &conf); err != nil {
		return nil, err
	}
	uploaderConf := defaultUploaderConfig
	if err := yaml.Unmarshal(uploaderBuf, &uploaderConf); err != nil {
		return nil, err
	}

	conf.DeviceName = uploaderConf.DeviceName

	if err := yaml.Unmarshal(locationBuf, &conf.Location); err != nil {
		return nil, err
	}

	if err := conf.Validate(); err != nil {
		return nil, err
	}

	return &conf, nil
}
