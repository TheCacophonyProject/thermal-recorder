// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2020, The Cacophony Project
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
)

type Config struct {
	DeviceID     int
	DeviceName   string
	FrameInput   string
	OutputDir    string
	MinDiskSpace uint64
}

func ParseConfig(configFolder string) (*Config, error) {
	configRW, err := goconfig.New(configFolder)
	if err != nil {
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
		DeviceID:   deviceConfig.ID,
		DeviceName: deviceConfig.Name,
		FrameInput: leptonConfig.FrameOutput,
		OutputDir:  "/var/spool/thermal-raw",
	}, nil
}
