// Copyright 2018 The Cacophony Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cptv

import (
	"bytes"
	"io"
	"testing"
	"time"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoundTripHeaderDefaults(t *testing.T) {
	cptvBytes := new(bytes.Buffer)

	w := NewWriter(cptvBytes)
	require.NoError(t, w.WriteHeader(Header{}))
	require.NoError(t, w.Close())

	r, err := NewReader(cptvBytes)
	require.NoError(t, err)
	assert.Equal(t, 2, r.Version())
	assert.True(t, time.Since(r.Timestamp()) < time.Minute) // "now" was used
	assert.Equal(t, "", r.DeviceName())
	assert.Equal(t, 0, r.PreviewSecs())
	assert.Equal(t, "", r.MotionConfig())
}

func TestRoundTripHeader(t *testing.T) {
	ts := time.Date(2016, 5, 4, 3, 2, 1, 0, time.UTC)
	cptvBytes := new(bytes.Buffer)

	w := NewWriter(cptvBytes)
	header := Header{
		Timestamp:    ts,
		DeviceName:   "nz42",
		PreviewSecs:  8,
		MotionConfig: "keep on movin",
	}
	require.NoError(t, w.WriteHeader(header))
	require.NoError(t, w.Close())

	r, err := NewReader(cptvBytes)
	require.NoError(t, err)
	assert.Equal(t, ts, r.Timestamp().UTC())
	assert.Equal(t, "nz42", r.DeviceName())
	assert.Equal(t, 8, r.PreviewSecs())
	assert.Equal(t, "keep on movin", r.MotionConfig())
}

func TestReaderFrameCount(t *testing.T) {
	frame := makeTestFrame()
	cptvBytes := new(bytes.Buffer)

	w := NewWriter(cptvBytes)
	require.NoError(t, w.WriteHeader(Header{}))
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.WriteFrame(frame))
	require.NoError(t, w.Close())

	r, err := NewReader(cptvBytes)
	require.NoError(t, err)
	c, err := r.FrameCount()
	require.NoError(t, err)
	assert.Equal(t, 3, c)
}

func TestFrameRoundTrip(t *testing.T) {
	frame0 := makeTestFrame()
	frame0.Status.TimeOn = 60 * time.Second
	frame0.Status.LastFFCTime = 30 * time.Second

	frame1 := makeOffsetFrame(frame0)
	frame1.Status.TimeOn = 61 * time.Second
	frame1.Status.LastFFCTime = 31 * time.Second

	frame2 := makeOffsetFrame(frame1)
	frame2.Status.TimeOn = 62 * time.Second
	frame2.Status.LastFFCTime = 32 * time.Second

	cptvBytes := new(bytes.Buffer)

	w := NewWriter(cptvBytes)
	require.NoError(t, w.WriteHeader(Header{}))
	require.NoError(t, w.WriteFrame(frame0))
	require.NoError(t, w.WriteFrame(frame1))
	require.NoError(t, w.WriteFrame(frame2))
	require.NoError(t, w.Close())

	r, err := NewReader(cptvBytes)
	require.NoError(t, err)

	frameD := new(lepton3.Frame)
	require.NoError(t, r.ReadFrame(frameD))
	assert.Equal(t, frame0, frameD)
	require.NoError(t, r.ReadFrame(frameD))
	assert.Equal(t, frame1, frameD)
	require.NoError(t, r.ReadFrame(frameD))
	assert.Equal(t, frame2, frameD)

	assert.Equal(t, io.EOF, r.ReadFrame(frameD))
}
