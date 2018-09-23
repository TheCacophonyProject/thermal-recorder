/*
audiobait - play sounds to lure animals for The Cacophony Project API.
Copyright (C) 2018, The Cacophony Project

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package window

import (
	"fmt"
	"time"
)

type TimeOfDay struct {
	time.Time
}

const timeLayout = `15:04`
const timeLayoutJson = `"` + timeLayout + `"`

func (timeOfDay *TimeOfDay) UnmarshalJSON(bValue []byte) error {
	sValue := string(bValue)
	if sValue == "null" {
		timeOfDay.Time = time.Time{}
		return nil
	}
	var err error
	timeOfDay.Time, err = time.Parse(timeLayoutJson, sValue)
	return err
}

func NewTimeOfDay(timeOfDayString string) *TimeOfDay {
	t, err := time.Parse(timeLayout, timeOfDayString)
	if err != nil {
		t = time.Time{}
	}
	return &TimeOfDay{Time: t}
}

func (timeOfDay *TimeOfDay) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var val string
	if err := unmarshal(&val); err != nil {
		return err
	}

	if val == "" {
		timeOfDay.Time = time.Time{}
		return nil
	}

	var err error
	timeOfDay.Time, err = time.Parse(timeLayout, val)
	if err != nil {
		return fmt.Errorf("Yaml error %v", err)
	}
	return err
}
