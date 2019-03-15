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
	"io"
	"time"

	"github.com/TheCacophonyProject/lepton3"
)

// NewWriter creates and returns a new Writer component
func NewWriter(w io.Writer) *Writer {
	return &Writer{
		bldr: NewBuilder(w),
		comp: NewCompressor(),
	}
}

// Writer uses a Builder and Compressor to create CPTV files.
type Writer struct {
	bldr *Builder
	comp *Compressor
}

// Header defines the information stored in the header of a CPTV
// file. All the fields are optional.
type Header struct {
	Timestamp    time.Time
	DeviceName   string
	PreviewSecs  int
	MotionConfig string
	Latitude     float32
	Longitude    float32
}

// WriteHeader writes a CPTV file header
func (w *Writer) WriteHeader(header Header) error {
	t := header.Timestamp
	if t.IsZero() {
		t = time.Now()
	}
	fields := NewFieldWriter()
	fields.Timestamp(Timestamp, t)
	fields.Uint32(XResolution, lepton3.FrameCols)
	fields.Uint32(YResolution, lepton3.FrameRows)
	fields.Uint8(Compression, 1)

	if len(header.DeviceName) > 0 {
		err := fields.String(DeviceName, header.DeviceName)
		if err != nil {
			return err
		}
	}

	fields.Uint8(PreviewSecs, uint8(header.PreviewSecs))

	if len(header.MotionConfig) > 0 {
		err := fields.String(MotionConfig, header.MotionConfig)
		if err != nil {
			return err
		}
	}

	if header.Latitude != 0.0 {
		fields.Float32(Latitude, header.Latitude)
	}
	if header.Longitude != 0.0 {
		fields.Float32(Longitude, header.Longitude)
	}

	return w.bldr.WriteHeader(fields)
}

// WriteFrame writes a CPTV frame
func (w *Writer) WriteFrame(frame *lepton3.Frame) error {
	bitWidth, compFrame := w.comp.Next(frame)
	fields := NewFieldWriter()
	fields.Uint32(TimeOn, durationToMillis(frame.Status.TimeOn))
	fields.Uint32(LastFFCTime, durationToMillis(frame.Status.LastFFCTime))
	fields.Uint8(BitWidth, uint8(bitWidth))
	fields.Uint32(FrameSize, uint32(len(compFrame)))
	return w.bldr.WriteFrame(fields, compFrame)
}

// Close closes the CPTV file
func (w *Writer) Close() error {
	return w.bldr.Close()
}

func durationToMillis(d time.Duration) uint32 {
	return uint32(d / time.Millisecond)
}
