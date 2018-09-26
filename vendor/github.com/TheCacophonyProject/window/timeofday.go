// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

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
	if err != nil {
		return fmt.Errorf("Json error %v", err)
	}
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
