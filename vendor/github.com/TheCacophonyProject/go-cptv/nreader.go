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
)

// nReader wraps an io.Reader providing some convenience methods for
// reading out specific lengths of data.
type nReader struct {
	io.Reader
}

// ReadByteInt reads a byte and returns it as an int type
func (r *nReader) ReadByteInt() (int, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return int(b), nil
}

// ReadByte reads a byte from the underlying reader
func (r *nReader) ReadByte() (byte, error) {
	buf, err := r.ReadN(1)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

// ReadN reads n bytes and returns as a byte array of size n.
func (r *nReader) ReadN(n int) ([]byte, error) {
	buf := make([]byte, n)
	i := 0
	for i < n {
		sz, err := r.Read(buf[i:])
		i += sz
		if err != nil {
			if err == io.EOF && i == n {
				// Read can return n,EOF from a read to the end of a file,
				// even when there was no attempt to go past the end. This
				// is normal io.Reader behavior. Because we're not returning
				// 'bytes read', we will break here and return err=nil in the
				// case where the requested number of bytes were sucessfully
				// read.
				break
			}
			return nil, err
		}
	}
	return buf, nil
}
