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
	"os"
)

// NewFileWriter creates file 'filename' and returns a new FileWriter
// with underlying buffer (bufio) Writer
func NewFileWriter(filename string) (*FileWriter, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	bw := bufio.NewWriter(f)
	return &FileWriter{
		Writer: NewWriter(bw),
		bw:     bw,
		f:      f,
	}, nil
}

// FileWriter wraps a Writer and provides a convenient way of writing
// a CPTV stream to a disk file.
type FileWriter struct {
	*Writer
	bw *bufio.Writer
	f  *os.File
}

// Name returns the name of the open File
func (fw *FileWriter) Name() string {
	return fw.f.Name()
}

// Close flushes the buffered writer and closes the open file
func (fw *FileWriter) Close() {
	fw.Writer.Close()
	fw.bw.Flush()
	fw.f.Close()
}
