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
	"testing"

	"github.com/TheCacophonyProject/window"
	"github.com/stretchr/testify/assert"
)

func TestWindowStartWithoutEndDoesntValidate(t *testing.T) {
	conf := RecorderConfig{
		WindowStart: *window.NewTimeOfDay("09:10"),
	}
	assert.EqualError(t, conf.Validate(), "window-start is set but window-end isn't")
}

func TestWindowEndWithoutStartDoesntValidate(t *testing.T) {
	conf := RecorderConfig{
		WindowEnd: *window.NewTimeOfDay("09:10"),
	}
	assert.EqualError(t, conf.Validate(), "window-end is set but window-start isn't")
}

func TestMinSecsGreaterThanMaxSecsDoesntValidate(t *testing.T) {
	conf := RecorderConfig{
		MinSecs: 5,
		MaxSecs: 2,
	}
	assert.EqualError(t, conf.Validate(), "max-secs should be larger than min-secs")
}
