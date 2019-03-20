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

	"github.com/TheCacophonyProject/window"
)

type RecorderConfig struct {
	MinSecs                int              `yaml:"min-secs"`
	MaxSecs                int              `yaml:"max-secs"`
	PreviewSecs            int              `yaml:"preview-secs"`
	UseSunriseSunsetWindow bool             `yaml:"sunrise-sunset"`
	SunriseOffset          int              `yaml:"sunrise-offset"`
	SunsetOffset           int              `yaml:"sunset-offset"`
	WindowStart            window.TimeOfDay `yaml:"window-start"`
	WindowEnd              window.TimeOfDay `yaml:"window-end"`
}

func DefaultRecorderConfig() RecorderConfig {
	return RecorderConfig{
		MinSecs:     10,
		MaxSecs:     600,
		PreviewSecs: 3,
	}
}

func (conf *RecorderConfig) Validate() error {
	if conf.MaxSecs < conf.MinSecs {
		return errors.New("max-secs should be larger than min-secs")
	}
	if conf.WindowStart.IsZero() && !conf.WindowEnd.IsZero() {
		return errors.New("window-end is set but window-start isn't")
	}
	if !conf.WindowStart.IsZero() && conf.WindowEnd.IsZero() {
		return errors.New("window-start is set but window-end isn't")
	}
	return nil
}
