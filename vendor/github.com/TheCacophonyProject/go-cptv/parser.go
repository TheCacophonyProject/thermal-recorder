package cptv

import (
	"compress/gzip"
	"errors"
	"fmt"
	"io"
)

// NewParser returns a new Parser instance, for parsing a gzip
// compressed CPTV stream using the provided Reader.
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

func (p *Parser) Frame() (Fields, []byte, error) {
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
	frameData, err := p.r.ReadN(int(frameSize))
	if err != nil {
		return nil, nil, err
	}
	return fields, frameData, nil
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
