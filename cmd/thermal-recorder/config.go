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
	goconfig "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/thermal-recorder/recorder"
	"github.com/TheCacophonyProject/thermal-recorder/throttle"
)

type Config struct {
	DeviceID     int
	DeviceName   string
	FrameInput   string
	OutputDir    string
	MinDiskSpace uint64
	Recorder     recorder.RecorderConfig
	Motion       goconfig.ThermalMotion
	Throttler    goconfig.ThermalThrottler
	Location     goconfig.Location
}

func ParseConfig(configFolder string) (*Config, error) {
	configRW, err := goconfig.New(configFolder)
	if err != nil {
		return nil, err
	}

	recorderConfig, err := recorder.NewConfig(configRW)
	if err != nil {
		return nil, err
	}

	motionConfig, err := motion.NewConfig(configRW)
	if err != nil {
		return nil, err
	}

	throttlerConfig, err := throttle.NewConfig(configRW)
	if err != nil {
		return nil, err
	}

	var locationConfig goconfig.Location
	if err := configRW.Unmarshal(goconfig.LocationKey, &locationConfig); err != nil {
		return nil, err
	}

	thermalRecorderConfig := goconfig.DefaultThermalRecorder()
	if err := configRW.Unmarshal(goconfig.ThermalRecorderKey, &thermalRecorderConfig); err != nil {
		return nil, err
	}

	leptonConfig := goconfig.DefaultLepton()
	if err := configRW.Unmarshal(goconfig.LeptonKey, &leptonConfig); err != nil {
		return nil, err
	}

	var deviceConfig goconfig.Device
	if err := configRW.Unmarshal(goconfig.DeviceKey, &deviceConfig); err != nil {
		return nil, err
	}

	return &Config{
		DeviceID:     deviceConfig.ID,
		DeviceName:   deviceConfig.Name,
		FrameInput:   leptonConfig.FrameOutput,
		OutputDir:    thermalRecorderConfig.OutputDir,
		MinDiskSpace: thermalRecorderConfig.MinDiskSpaceMB,
		Recorder:     *recorderConfig,
		Motion:       *motionConfig,
		Throttler:    *throttlerConfig,
		Location:     locationConfig,
	}, nil
}
