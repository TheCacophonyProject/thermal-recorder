// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

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
