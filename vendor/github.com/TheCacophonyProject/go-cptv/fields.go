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
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"time"
)

// ReadFields reads the fields for a CPTV section, returning a new
// Fields instance.
func ReadFields(r io.Reader) (Fields, error) {
	return readFieldsN(nReader{r})
}

func readFieldsN(r nReader) (Fields, error) {
	fieldCount, err := r.ReadByteInt()
	if err != nil {
		return nil, err
	}
	f := make(Fields)
	for i := 0; i < fieldCount; i++ {
		size, err := r.ReadByteInt()
		if err != nil {
			return nil, err
		}
		code, err := r.ReadByte()
		if err != nil {
			return nil, err
		}
		data, err := r.ReadN(size)
		if err != nil {
			return nil, err
		}
		f[code] = data
	}
	return f, nil
}

// Fields maps from field key -> field data
type Fields map[byte][]byte

// Uint8 returns the field at 'key' as a uint8
func (f Fields) Uint8(key byte) (uint8, error) {
	buf, err := f.get(key, 1)
	if err != nil {
		return 0, err
	}
	return uint8(buf[0]), nil
}

// Uint16 returns the field at 'key' as a uint16
func (f Fields) Uint16(key byte) (uint16, error) {
	buf, err := f.get(key, 2)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint16(buf), nil
}

// Uint32 returns the field at 'key' as a uint32
func (f Fields) Uint32(key byte) (uint32, error) {
	buf, err := f.get(key, 4)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint32(buf), nil
}

// Uint64 returns the field at 'key' as a uint64
func (f Fields) Uint64(key byte) (uint64, error) {
	buf, err := f.get(key, 8)
	if err != nil {
		return 0, err
	}
	return binary.LittleEndian.Uint64(buf), nil
}

// Timestamp returns the field at 'key' as a time value
func (f Fields) Timestamp(key byte) (time.Time, error) {
	tRaw, err := f.Uint64(key)
	if err != nil {
		return time.Time{}, err
	}
	return time.Unix(0, int64(tRaw*1000)), nil
}

// String returns the field at 'key' as a character string
func (f Fields) String(key byte) (string, error) {
	buf, ok := f[key]
	if !ok {
		return "", errors.New("not found")
	}
	return string(buf), nil
}

// get returns the field at 'key' as a byte array after checking
// its size vs expectedLen
func (f Fields) get(key byte, expectedLen int) ([]byte, error) {
	buf, ok := f[key]
	if !ok {
		return nil, errors.New("not found")
	}
	if len(buf) != expectedLen {
		return nil, fmt.Errorf("expected length %d, got %d", expectedLen, len(buf))
	}
	return buf, nil
}

// NewFieldWriter creates a new FieldWriter
func NewFieldWriter() *FieldWriter {
	return &FieldWriter{
		data: make([]byte, 0, 128),
	}
}

// FieldWriter generates CPTV encoded fields.
type FieldWriter struct {
	data       []byte
	fieldCount uint8
}

// Uint8 writes a uint8 field with key 'code' and value 'v'
func (f *FieldWriter) Uint8(code byte, v uint8) {
	f.data = append(f.data, byte(1), code, byte(v))
	f.fieldCount++
}

// Uint16 writes a uint16 field with key 'code' and value 'v'
func (f *FieldWriter) Uint16(code byte, v uint16) {
	b := []byte{2, code, 0, 0}
	binary.LittleEndian.PutUint16(b[2:], v)
	f.data = append(f.data, b...)
	f.fieldCount++
}

// Uint32 writes a uint32 field with key 'code' and value 'v'
func (f *FieldWriter) Uint32(code byte, v uint32) {
	b := []byte{4, code, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(b[2:], v)
	f.data = append(f.data, b...)
	f.fieldCount++
}

// Uint64 writes a uint64 field with key 'code' and value 'v'
func (f *FieldWriter) Uint64(code byte, v uint64) {
	b := []byte{8, code, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint64(b[2:], v)
	f.data = append(f.data, b...)
	f.fieldCount++
}

// Timestamp writes a time field with key 'code' and value 'v'
func (f *FieldWriter) Timestamp(code byte, t time.Time) {
	f.Uint64(code, uint64(t.UnixNano()/1000))
}

// String writes a character string field with key 'code' and value 'v'
func (f *FieldWriter) String(code byte, v string) error {
	if len(v) > 255 {
		return fmt.Errorf("string length %d greater than 255", len(v))
	}

	byteSlice := []byte(v)
	lenByteSlice := byte(len(byteSlice))
	f.data = append(f.data, lenByteSlice)
	f.data = append(f.data, code)
	f.data = append(f.data, byteSlice...)
	f.fieldCount++

	return nil
}
