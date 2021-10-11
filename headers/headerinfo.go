// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2020, The Cacophony Project
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

package headers

import (
	"bufio"
	"bytes"
	"strings"

	"gopkg.in/yaml.v1"
)

// HeaderInfo contains the camera description fields returned by a
// camera service.
type HeaderInfo struct {
	resX      int
	resY      int
	fps       int
	framesize int
	brand     string
	model     string
	serial    int
	firmware  string
}

// ResX implements cptvframe.CameraSpec.
func (h *HeaderInfo) ResX() int {
	return h.resX
}

// ResY implements cptvframe.CameraSpec.
func (h *HeaderInfo) ResY() int {
	return h.resY
}

// FPS implements cptvframe.CameraSpec.
func (h *HeaderInfo) FPS() int {
	return h.fps
}

// FrameSize returns the number of bytes in each frame (include any
// telemetry bytes).
func (h *HeaderInfo) FrameSize() int {
	return h.framesize
}

// Model returns the camera model.
func (h *HeaderInfo) Model() string {
	return h.model
}

// Brand returns the camera brand.
func (h *HeaderInfo) Brand() string {
	return h.brand
}

// Camera module firmware revision i.e. 1.2.3
func (h *HeaderInfo) Firmware() string {
	return h.firmware
}

// Camera module unique serial#
func (h *HeaderInfo) CameraSerial() int {
	return h.serial
}

func ReadHeaderInfo(reader *bufio.Reader) (*HeaderInfo, error) {
	var buf bytes.Buffer
	for {
		line, err := reader.ReadString(byte('\n'))
		if err != nil {
			return nil, err
		}
		if strings.Trim(line, " ") == "\n" {
			break
		}
		buf.WriteString(line)
	}
	h := make(map[string]interface{})
	err := yaml.Unmarshal(buf.Bytes(), &h)
	if err != nil {
		return nil, err
	}
	header := &HeaderInfo{
		resX:      toInt(h[XResolution]),
		resY:      toInt(h[YResolution]),
		fps:       toInt(h[FPS]),
		framesize: toInt(h[FrameSize]),
		brand:     toStr(h[Brand]),
		model:     toStr(h[Model]),
		serial:    toInt(h[Serial]),
		firmware:  toStr(h[Firmware]),
	}
	return header, nil
}

func toInt(v interface{}) int {
	out, ok := v.(int)
	if !ok {
		return 0
	}
	return out
}

func toStr(v interface{}) string {
	out, ok := v.(string)
	if !ok {
		return ""
	}
	return out
}
