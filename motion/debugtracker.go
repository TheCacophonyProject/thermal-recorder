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
	"fmt"
	"math"
	"strings"
)

type debugTracker struct {
	values map[string]*value
}

func newDebugTracker() *debugTracker {
	return &debugTracker{
		values: make(map[string]*value),
	}
}

func (d *debugTracker) update(name string, x int) {
	if d == nil {
		return
	}
	value := d.values[name]
	if value == nil {
		value = newValue()
		d.values[name] = value
	}
	value.update(x)
}

func (d *debugTracker) reset() {
	if d == nil {
		return
	}
	for _, value := range d.values {
		value.reset()
	}
}

// string takes a format string and generates a string representation
// of the tracked values.  A format string looks like "foo:min
// bar:all" where foo and bar are field names and min and all are
// output formats.
// Available formats are "n", "min", "max", "avg", "all".
func (d *debugTracker) string(format string) string {
	if d == nil {
		return ""
	}
	var out []string
	for _, field := range strings.Split(format, " ") {
		parts := strings.Split(field, ":")
		if len(parts) != 2 {
			continue
		}
		name, style := parts[0], parts[1]
		value := d.values[name]
		if value != nil {
			out = append(out, fmt.Sprintf("%s: %s", name, value.format(style)))
		}
	}
	return strings.Join(out, "; ")
}

func newValue() *value {
	v := new(value)
	v.reset()
	return v
}

type value struct {
	n   int
	min int
	max int
	avg float64
}

func (v *value) reset() {
	v.n = 0
	v.max = math.MinInt32
	v.min = math.MaxInt32
	v.avg = 0
}

func (v *value) update(x int) {
	v.n++
	if x > v.max {
		v.max = x
	}
	if x < v.min {
		v.min = x
	}
	// Cumulative moving average
	v.avg = v.avg + ((float64(x) - v.avg) / float64(v.n))
}

func (v *value) format(style string) string {
	switch style {
	case "n":
		return fmt.Sprint(v.n)
	case "min":
		return fmt.Sprint(v.min) + "(min)"
	case "max":
		return fmt.Sprint(v.max) + "(max)"
	case "avg":
		return fmt.Sprintf("%.2f(avg)", v.avg)
	case "all":
		return fmt.Sprintf("%d -> %d (avg: %.2f)", v.min, v.max, v.avg)
	default:
		return "???"
	}
}
