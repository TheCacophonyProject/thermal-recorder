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

import (
	"errors"
	"time"

	yaml "gopkg.in/yaml.v2"
)

const (
	defaultConfig = "/etc/cacophony/location.yaml"
	maxLatitude   = 90
	maxLongitude  = 180

	//Christchurch
	defaultLatitude  = -43.5321
	defaultLongitude = 172.6362

	defaultAltitude = 0
	defaultAccuracy = 10
)

type LocationConfig struct {
	Latitude     float32   `yaml:"latitude"`
	Longitude    float32   `yaml:"longitude"`
	LocTimestamp time.Time `yaml:"timestamp"`
	Altitude     float32   `yaml:"altitude"`
	Accuracy     float32   `yaml:"accuracy"`
}

func DefaultLocationFile() string {
	return defaultConfig
}

func DefaultLocationConfig() LocationConfig {
	return LocationConfig{
		Latitude:     defaultLatitude,
		Longitude:    defaultLongitude,
		LocTimestamp: time.Time{},
		Altitude:     defaultAltitude,
		Accuracy:     defaultAccuracy,
	}
}

func (conf *LocationConfig) IsLocationEmpty() bool {
	return conf.Latitude == 0 && conf.Longitude == 0
}

//ParseConfig stored in yaml bytes and sets default location if location is (0,0)
func (conf *LocationConfig) ParseConfig(buf []byte) error {
	err := yaml.Unmarshal(buf, conf)
	if err == nil && conf.IsLocationEmpty() {
		conf.Latitude = defaultLatitude
		conf.Longitude = defaultLongitude
		conf.LocTimestamp = time.Time{}
		conf.Altitude = defaultAltitude
		conf.Accuracy = defaultAccuracy
	}

	return err
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

	// NB: The LocTimestamp, Altitude and Accuracy fields are
	//     a) Optional and
	//     b) Set by the management interface. Their values are checked at that time.

	return nil
}
