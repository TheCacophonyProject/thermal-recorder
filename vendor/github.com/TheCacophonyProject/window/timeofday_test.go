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
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestTimeOfDay struct {
	Time TimeOfDay
}

func TestParsingValidTimeJson(t *testing.T) {
	parseJsonTimeAndCheck(t, "2:03", "02:03")
	parseJsonTimeAndCheck(t, "0:00", "00:00")
	parseJsonTimeAndCheck(t, "21:55", "21:55")
	parseJsonShouldFail(t, "25:01")
	parseJsonShouldFail(t, "20:67")

}

func parseJsonTimeAndCheck(t *testing.T, time string, checktime string) {
	timeOfDay, err := parseJsonTime(time)

	if err != nil {
		t.Errorf("Unexpected error unmarshalling: %s", err)
	}

	outputTime := fmt.Sprintf("%02d:%02d", timeOfDay.Time.Hour(), timeOfDay.Time.Minute())

	assert.Equal(t, outputTime, checktime)

	fmt.Printf("Unmarshalled as %s\n", outputTime)
}

func parseJsonShouldFail(t *testing.T, time string) {
	if _, err := parseJsonTime(time); err == nil {
		t.Errorf("Should not have parsed time correctly: %s", time)
	} else {
		fmt.Println("Invalid time didn't parse (this is the expected result).")
	}
}

func parseJsonTime(time string) (TestTimeOfDay, error) {
	var timeOfDay TestTimeOfDay
	fmt.Printf("Parsing time '%s'.\n", time)

	data := []byte(fmt.Sprintf(`{"time": "%s"}`, time))

	err := json.Unmarshal(data, &timeOfDay)

	return timeOfDay, err
}
