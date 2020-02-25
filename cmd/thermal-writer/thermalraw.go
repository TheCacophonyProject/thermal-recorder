// Copyright 2020 The Cacophony Project
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

package main

import (
	"fmt"
	"io"
	"log"
	"path/filepath"
	"time"

	"github.com/TheCacophonyProject/go-cptv"

	"github.com/TheCacophonyProject/thermal-recorder/headers"
)

const (
	thermalRawMagic        = "CPTR"
	thermalRawVersion byte = 0x02

	headerSection = 'H'
	frameSection  = 'F'
)

func newThermalRaw(conf *Config, t time.Time, h *headers.HeaderInfo) (*Builder, error) {
	f, err := nextFile(conf.OutputDir)
	if err != nil {
		return nil, err
	}
	b := newBuilder(f)

	fields := cptv.NewFieldWriter()
	fields.Timestamp(cptv.Timestamp, t)
	fields.String(cptv.Model, h.Model())
	fields.String(cptv.Brand, h.Brand())
	fields.Uint8(cptv.FPS, uint8(h.FPS()))
	fields.Uint32(cptv.XResolution, uint32(h.ResX()))
	fields.Uint32(cptv.YResolution, uint32(h.ResY()))
	fields.Uint8(cptv.Compression, 0)
	fields.String(cptv.DeviceName, conf.DeviceName)
	fields.Uint32(cptv.DeviceID, uint32(conf.DeviceID))
	if err := b.WriteHeader(fields); err != nil {
		return nil, err
	}

	return b, nil
}

func writeFrame(b *Builder, frame []byte) error {
	fields := cptv.NewFieldWriter()
	fields.Uint32(cptv.FrameSize, uint32(len(frame)))
	return b.WriteFrame(fields, frame)
}

func nextFile(outDir string) (*bufferedFile, error) {
	n := nextFileName(outDir)
	log.Println("writing to", n)
	return newBufferedFile(n)
}

func nextFileName(outDir string) string {
	name := fmt.Sprintf("%s.thermalraw", time.Now().Format("2006_01_02T15_04_05"))
	return filepath.Join(outDir, name)
}

// newBuilder returns a new Builder instance, ready to generate a raw
// thermal file.
func newBuilder(w io.WriteCloser) *Builder {
	return &Builder{
		w: w,
	}
}

// Builder handles the low-level construction of thermal raw sections
// and fields.
type Builder struct {
	w io.WriteCloser
}

func (b *Builder) WriteHeader(f *cptv.FieldWriter) error {
	fieldData, numFields := f.Bytes()
	_, err := b.w.Write(append(
		[]byte(thermalRawMagic),
		thermalRawVersion,
		headerSection,
		byte(numFields),
	))
	if err != nil {
		return err
	}

	_, err = b.w.Write(fieldData)
	return err
}

func (b *Builder) WriteFrame(f *cptv.FieldWriter, frameData []byte) error {
	// Frame header
	fieldData, numFields := f.Bytes()
	_, err := b.w.Write([]byte{frameSection, byte(numFields)})
	if err != nil {
		return err
	}

	// Frame fields
	_, err = b.w.Write(fieldData)
	if err != nil {
		return err
	}

	// Frame thermal data
	_, err = b.w.Write(frameData)
	return err
}

func (b *Builder) Close() error {
	return b.w.Close()
}
