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

// NewFileReader returns a new FileReader from the filename.
func NewFileReader(filename string) (*FileReader, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	br := bufio.NewReader(f)
	r, err := NewReader(br)
	if err != nil {
		f.Close()
		return nil, err
	}
	return &FileReader{
		Reader: r,
		br:     br,
		f:      f,
	}, nil
}

// FileReader wraps a Reader and provides a convenient way of reading
// a CPTV stream from a disk file.
type FileReader struct {
	*Reader
	br *bufio.Reader
	f  *os.File
}

// Name returns the name of the FileReader
func (fr *FileReader) Name() string {
	return fr.f.Name()
}

// Close closes the file
func (fr *FileReader) Close() {
	fr.f.Close()
}
