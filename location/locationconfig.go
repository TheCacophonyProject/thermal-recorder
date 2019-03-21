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

package location

import "errors"

const (
	defaultConfig = "/etc/cacophony/location.yaml"
	maxLatitude   = 90
	maxLongitude  = 180
)

type LocationConfig struct {
	Latitude  float32 `yaml:"latitude"`
	Longitude float32 `yaml:"longitude"`
}

func DefaultLocationFile() string {
	return defaultConfig
}

func DefaultLocationConfig() LocationConfig {
	return LocationConfig{
		Latitude:  -43.5321,
		Longitude: 172.6362,
	}
}
func (conf *LocationConfig) Validate() error {
	if &conf.Latitude == nil {
		return errors.New("Latitude cannot be nil")
	}
	if conf.Latitude < -maxLatitude || conf.Latitude > maxLatitude {
		return errors.New("Latitude outside of normal range")
	}

	if &conf.Longitude == nil {
		return errors.New("Longitude cannot be nil")
	}
	if conf.Longitude < -maxLongitude || conf.Longitude > maxLongitude {
		return errors.New("Longitude outside of normal range")
	}
	return nil
}
