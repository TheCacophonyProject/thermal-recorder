// thermal-recorder - record thermal video footage of warm moving objects
// Copyright (C) 2019, The Cacophony Project
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

package loglimiter

import (
	"bytes"
	"log"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestPrint(t *testing.T) {
	logs, reset := captureLogs()
	defer reset()

	limiter := New(time.Minute)
	limiter.Print("hello")
	limiter.Print("world")

	assert.Equal(t, "hello\nworld\n", logs.String())
}

func TestPrintf(t *testing.T) {
	logs, reset := captureLogs()
	defer reset()

	limiter := New(time.Minute)
	limiter.Printf("hello: %d", 42)
	limiter.Printf("world: %q", "hi")

	assert.Equal(t, "hello: 42\nworld: \"hi\"\n", logs.String())
}

func TestLimitPrint(t *testing.T) {
	logs, reset := captureLogs()
	defer reset()

	now := time.Now()

	limiter := New(2 * time.Second)
	limiter.nowFunc = func() time.Time { return now }

	limiter.Print("hello")
	assert.Equal(t, "hello\n", logs.String())

	// Advance time but still within the window.
	now = now.Add(time.Second)
	limiter.Print("hello")
	assert.Equal(t, "hello\n", logs.String())

	// Now go past the window; see that second line is logged.
	now = now.Add(time.Second)
	limiter.Print("hello")
	assert.Equal(t, "hello\nhello\n", logs.String())

	// Log something else and see that this is let through.
	limiter.Print("world")
	assert.Equal(t, "hello\nhello\nworld\n", logs.String())

	// Log again, and see it be suppressed..
	limiter.Print("world")
	assert.Equal(t, "hello\nhello\nworld\n", logs.String())
}

func TestLimitPrintf(t *testing.T) {
	logs, reset := captureLogs()
	defer reset()

	now := time.Now()

	limiter := New(2 * time.Second)
	limiter.nowFunc = func() time.Time { return now }

	limiter.Printf("hello")
	assert.Equal(t, "hello\n", logs.String())

	// Advance time but still within the window.
	now = now.Add(time.Second)
	limiter.Printf("hello")
	assert.Equal(t, "hello\n", logs.String())

	// Now go past the window; see that second line is logged.
	now = now.Add(time.Second)
	limiter.Printf("hello")
	assert.Equal(t, "hello\nhello\n", logs.String())
}

func TestMixed(t *testing.T) {
	logs, reset := captureLogs()
	defer reset()

	// Mixing Print and Printf doesn't matter if the resulting string is the same.
	limiter := New(time.Minute)
	limiter.Print("hello")
	limiter.Printf("hello")
	assert.Equal(t, "hello\n", logs.String())
}

func captureLogs() (*bytes.Buffer, func()) {
	flags := log.Flags()
	log.SetFlags(0)

	logs := new(bytes.Buffer)
	log.SetOutput(logs)

	return logs, func() {
		log.SetOutput(os.Stderr)
		log.SetFlags(flags)
	}
}
