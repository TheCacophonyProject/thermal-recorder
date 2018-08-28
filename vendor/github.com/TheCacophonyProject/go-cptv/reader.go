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
