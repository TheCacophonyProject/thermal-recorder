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

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAllDefaults(t *testing.T) {
	conf, err := ParseConfig([]byte(""))
	require.NoError(t, err)

	assert.Equal(t, Config{
		SPISpeed:    2000000,
		PowerPin:    "GPIO23",
		FrameOutput: "/var/run/lepton-frames",
	}, *conf)
}

func TestAllSet(t *testing.T) {
	// All config set at non-default values.
	config := []byte(`
spi-speed: 123
power-pin: "PIN"
frame-output: "/some/sock"
`)

	conf, err := ParseConfig(config)
	require.NoError(t, err)

	assert.Equal(t, Config{
		SPISpeed:    123,
		PowerPin:    "PIN",
		FrameOutput: "/some/sock",
	}, *conf)
}
