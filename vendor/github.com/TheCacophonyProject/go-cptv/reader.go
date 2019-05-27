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
	"bufio"
	"io"
	"io/ioutil"
	"time"

	"github.com/TheCacophonyProject/lepton3"
)

// NewReader returns a new Reader from the io.Reader given.
func NewReader(r io.Reader) (*Reader, error) {
	parser, err := NewParser(bufio.NewReader(r))
	if err != nil {
		return nil, err
	}
	header, err := parser.Header()
	if err != nil {
		return nil, err
	}
	return &Reader{
		parser: parser,
		decomp: NewDecompressor(),
		header: header,
	}, nil
}

// Reader uses a Parser and Decompressor to read CPTV recordings.
type Reader struct {
	parser *Parser
	decomp *Decompressor
	header Fields
}

// Version returns the version number of the CPTV file.
func (r *Reader) Version() int {
	return r.parser.version
}

// Timestamp returns the CPTV timestamp. A zero time is returned if
// the field wasn't present (shouldn't happen).
func (r *Reader) Timestamp() time.Time {
	ts, _ := r.header.Timestamp(Timestamp)
	return ts
}

// DeviceName returns the device name field from the CPTV
// recording. Returns an empty string if the device name field wasn't
// present.
func (r *Reader) DeviceName() string {
	name, _ := r.header.String(DeviceName)
	return name
}

// PreviewSecs returns the number of seconds included in the recording
// before motion was detected. Returns 0 if this field is not included.
func (r *Reader) PreviewSecs() int {
	secs, _ := r.header.Uint8(PreviewSecs)
	return int(secs)
}

// MotionConfig returns the YAML configuration for the motion detector
// that was in use when this CPTV file was recorded. Returns an empty string
// if this field is not included.
func (r *Reader) MotionConfig() string {
	conf, _ := r.header.String(MotionConfig)
	return conf
}

// Latitude returns the latitude part of the location of the device
// when this CPTV file was recorded. Returns 0 if the field is not included.
func (r *Reader) Latitude() float32 {
	lat, _ := r.header.Float32(Latitude)
	return lat
}

// Longitude returns the longitude part of the location of the device
// when this CPTV file was recorded. Returns 0 if the field is not included.
func (r *Reader) Longitude() float32 {
	long, _ := r.header.Float32(Longitude)
	return long
}

// LocTimestamp returns the timestamp at which the location of the device.
// Returns the nil time.Time value if the field is not included.
func (r *Reader) LocTimestamp() time.Time {
	ts, _ := r.header.Timestamp(LocTimestamp)
	return ts
}

// Altitude returns the altitude part of the location of the device
// when this CPTV file was recorded. Returns 0 if the field is not included.
func (r *Reader) Altitude() float32 {
	alt, _ := r.header.Float32(Altitude)
	return alt
}

// Accuracy returns the estimated accuracy of the location setting of the device
// when this CPTV file was recorded. Returns 0 if the field is not included.
func (r *Reader) Accuracy() float32 {
	pre, _ := r.header.Float32(Accuracy)
	return pre
}

// ReadFrame extracts and decompresses the next frame in a CPTV
// recording. At the end of the recording an io.EOF error will be
// returned.
func (r *Reader) ReadFrame(out *lepton3.Frame) error {
	fields, frameReader, err := r.parser.Frame()
	if err != nil {
		return err
	}
	bitWidth, err := fields.Uint8(BitWidth)
	if err != nil {
		return err
	}

	// This field is garbage below v2 so ignore it for older files.
	if r.parser.version >= 2 {
		timeOn, err := fields.Uint32(TimeOn)
		if err == nil {
			out.Status.TimeOn = millisToDuration(timeOn)
		}
	}

	lastFFCTime, err := fields.Uint32(LastFFCTime)
	if err == nil {
		out.Status.LastFFCTime = millisToDuration(lastFFCTime)
	}

	return r.decomp.Next(bitWidth, &nReader{frameReader}, out)
}

// FrameCount returns the remaining number of frames in a CPTV file.
// After this call, all remaining frames will have been consumed.
func (r *Reader) FrameCount() (int, error) {
	count := 0
	for {
		_, fr, err := r.parser.Frame()
		if err != nil {
			if err == io.EOF {
				break
			}
			return count, err
		}
		io.Copy(ioutil.Discard, fr)
		count++
	}
	return count, nil
}

func millisToDuration(ms uint32) time.Duration {
	return time.Duration(ms) * time.Millisecond
}
