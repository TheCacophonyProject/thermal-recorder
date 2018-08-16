package cptv

import (
	"compress/gzip"
	"io"
)

// NewBuilder returns a new Builder instance, ready to emit a gzip
// compressed CPTV file to the provided Writer.
func NewBuilder(w io.Writer) *Builder {
	return &Builder{
		w: gzip.NewWriter(w),
	}
}

// Builder handles the low-level construction of CPTV sections and
// fields. See Writer for a higher-level interface.
type Builder struct {
	w *gzip.Writer
}

func (b *Builder) WriteHeader(f *FieldWriter) error {
	_, err := b.w.Write(append(
		[]byte(magic),
		version,
		headerSection,
		byte(f.fieldCount),
	))
	if err != nil {
		return err
	}

	_, err = b.w.Write(f.data)
	return err
}

func (b *Builder) WriteFrame(f *FieldWriter, frameData []byte) error {
	// Frame header
	_, err := b.w.Write([]byte{frameSection, byte(f.fieldCount)})
	if err != nil {
		return err
	}

	// Frame fields
	_, err = b.w.Write(f.data)
	if err != nil {
		return err
	}

	// Frame thermal data
	_, err = b.w.Write(frameData)
	return err
}

func (b *Builder) Close() error {
	if err := b.w.Flush(); err != nil {
		return err
	}
	return b.w.Close()
}
