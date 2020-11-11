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
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
)

type Recorder interface {
	StopRecording() error
	StartRecording(backgroundFrame *cptvframe.Frame, tempThresh uint16) error
	WriteFrame(frame *cptvframe.Frame) error
	CheckCanRecord() error
}

type NoWriteRecorder struct {
}

func (*NoWriteRecorder) StopRecording() error { return nil }
func (*NoWriteRecorder) StartRecording(backgroundFrame *cptvframe.Frame, tempThresh uint16) error {
	return nil
}
func (*NoWriteRecorder) WriteFrame(frame *cptvframe.Frame) error { return nil }
func (*NoWriteRecorder) CheckCanRecord() error                   { return nil }
