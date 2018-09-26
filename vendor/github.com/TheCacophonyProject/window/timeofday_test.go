// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package window

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

type TestTimeOfDay struct {
	Time TimeOfDay
}

func TestParsingTimeJson(t *testing.T) {
	parseJsonTimeAndCheck(t, "2:03", "02:03")
	parseJsonTimeAndCheck(t, "0:00", "00:00")
	parseJsonTimeAndCheck(t, "21:55", "21:55")

	_, err := parseJsonTime("25:01")
	assert.EqualError(t, err, `Json error parsing time ""25:01"": hour out of range`)

	_, err = parseJsonTime("20:67")
	assert.EqualError(t, err, `Json error parsing time ""20:67"": minute out of range`)

}

func TestParsingTimeYaml(t *testing.T) {
	parseYamlTimeAndCheck(t, "02:03")
	_, err := parseYamlTime("02:62")
	assert.EqualError(t, err, `Yaml error parsing time "02:62": minute out of range`)
}

func parseJsonTimeAndCheck(t *testing.T, time string, checktime string) {
	timeOfDay, err := parseJsonTime(time)
	if err != nil {
		t.Errorf("Unexpected error unmarshalling: %s", err)
	}
	outputTime := fmt.Sprintf("%02d:%02d", timeOfDay.Time.Hour(), timeOfDay.Time.Minute())
	assert.Equal(t, checktime, outputTime)
}

func parseJsonTime(time string) (TestTimeOfDay, error) {
	var timeOfDay TestTimeOfDay
	err := json.Unmarshal([]byte(`{"time": "`+time+`"}`), &timeOfDay)
	return timeOfDay, err
}

func parseYamlTimeAndCheck(t *testing.T, time string) {
	timeOfDay, err := parseYamlTime(time)

	if err != nil {
		t.Errorf("Unexpected error unmarshalling: %s", err)
	}

	fmt.Printf("%v", timeOfDay)
	outputTime := fmt.Sprintf("%02d:%02d", timeOfDay.Time.Hour(), timeOfDay.Time.Minute())
	assert.Equal(t, time, outputTime)
}

func parseYamlTime(time string) (TestTimeOfDay, error) {
	timeOfDay := TestTimeOfDay{}
	err := yaml.Unmarshal([]byte(`"time": `+time), &timeOfDay)
	return timeOfDay, err
}
