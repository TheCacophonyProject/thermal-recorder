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
	"compress/gzip"
	"errors"
	"fmt"
	"io"
)

// NewParser returns a new Parser instance, for parsing a gzip
// compressed CPTV stream using the provided io.Reader.
//
// Providing a buffered Reader is preferable.
func NewParser(r io.Reader) (*Parser, error) {
	gr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}
	return &Parser{
		r: nReader{gr},
	}, nil
}

// Parser is the low-level type for pulling apart the sections and
// fields of a CPTV file. See Reader for a high-level interface.
type Parser struct {
	r nReader
}

func (p *Parser) Header() (Fields, error) {
	if magicRead, err := p.r.ReadN(4); err != nil {
		return nil, err
	} else if string(magicRead) != magic {
		return nil, errors.New("magic not found")
	}

	if err := p.checkByte("version", version); err != nil {
		return nil, err
	}

	if err := p.checkByte("section", headerSection); err != nil {
		return nil, err
	}

	return readFieldsN(p.r)
}

func (p *Parser) Frame() (Fields, io.Reader, error) {
	if err := p.checkByte("section", frameSection); err != nil {
		return nil, nil, err
	}
	fields, err := readFieldsN(p.r)
	if err != nil {
		return nil, nil, err
	}
	frameSize, err := fields.Uint32(FrameSize)
	if err != nil {
		return nil, nil, err
	}

	// Return a subreader which is only allows access to the bytes for
	// the frame.
	frameReader := &io.LimitedReader{
		R: p.r.Reader,
		N: int64(frameSize),
	}
	return fields, frameReader, nil
}

func (p *Parser) checkByte(label string, expected byte) error {
	actual, err := p.r.ReadByte()
	if err != nil {
		return err
	}
	if actual != expected {
		return fmt.Errorf("unexpected %s: %d", label, actual)
	}
	return nil
}
