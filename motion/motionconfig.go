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

package motion

import (
	config "github.com/TheCacophonyProject/go-config"
)

func NewConfig(configRW *config.Config, cameraModel string) (*config.ThermalMotion, error) {
	thermalMotionConfig := config.DefaultThermalMotion(cameraModel)
	if err := configRW.Unmarshal(config.ThermalMotionKey, &thermalMotionConfig); err != nil {
		return nil, err
	}
	if err := validateConfig(&thermalMotionConfig); err != nil {
		return nil, err
	}
	return &thermalMotionConfig, nil
}

func validateConfig(*config.ThermalMotion) error {
	// TODO
	return nil
}
