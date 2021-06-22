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

package recorder

import (
	"errors"

	config "github.com/TheCacophonyProject/go-config"
	"github.com/TheCacophonyProject/window"
)

type RecorderConfig struct {
	MinSecs          int
	MaxSecs          int
	PreviewSecs      int
	Window           window.Window
	ConstantRecorder bool
}

func NewConfig(conf *config.Config) (*RecorderConfig, error) {
	thermalRecorderConfig := config.DefaultThermalRecorder()
	if err := conf.Unmarshal(config.ThermalRecorderKey, &thermalRecorderConfig); err != nil {
		return nil, err
	}
	windowLocationConfig := config.DefaultWindowLocation()
	if err := conf.Unmarshal(config.LocationKey, &windowLocationConfig); err != nil {
		return nil, err
	}
	windowsConfig := config.DefaultWindows()
	if err := conf.Unmarshal(config.WindowsKey, &windowsConfig); err != nil {
		return nil, err
	}

	w, err := window.New(
		windowsConfig.StartRecording,
		windowsConfig.StopRecording,
		float64(windowLocationConfig.Latitude),
		float64(windowLocationConfig.Longitude))
	if err != nil {
		return nil, err
	}

	recorderConfig := RecorderConfig{
		MinSecs:          thermalRecorderConfig.MinSecs,
		MaxSecs:          thermalRecorderConfig.MaxSecs,
		PreviewSecs:      thermalRecorderConfig.PreviewSecs,
		Window:           *w,
		ConstantRecorder: thermalRecorderConfig.ConstantRecorder,
	}

	if err := recorderConfig.validate(); err != nil {
		return nil, err
	}
	return &recorderConfig, nil
}

func (conf *RecorderConfig) validate() error {
	if conf.MaxSecs < conf.MinSecs {
		return errors.New("max-secs should be larger than min-secs")
	}
	return nil
}
