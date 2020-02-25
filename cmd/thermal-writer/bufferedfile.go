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

package main

import (
	"bufio"
	"os"
)

func newBufferedFile(filename string) (*bufferedFile, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return &bufferedFile{
		f: f,
		w: bufio.NewWriterSize(f, 32*1024*1024),
	}, nil
}

type bufferedFile struct {
	f *os.File
	w *bufio.Writer
}

func (bf *bufferedFile) Write(p []byte) (int, error) {
	return bf.w.Write(p)
}

func (bf *bufferedFile) Close() error {
	if err := bf.w.Flush(); err != nil {
		return err
	}
	return bf.f.Close()
}
