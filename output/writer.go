package output

import (
	"compress/gzip"
	"encoding/binary"
	"io"
	"time"
)

const (
	headerCode = 'H'
	frameCode  = 'F'

	// Header fields
	Timestamp   byte = 'T'
	XResolution byte = 'X'
	YResolution byte = 'Y'
	Compression byte = 'C'

	// Frame fields
	Offset    byte = 't'
	BitWidth  byte = 'w'
	FrameSize byte = 'f'

	magic        = "CPTV"
	version byte = 0x01
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		w: gzip.NewWriter(w),
	}
}

type Writer struct {
	w *gzip.Writer
}

func (w *Writer) WriteHeader(f *Fields) error {
	_, err := w.w.Write(append(
		[]byte(magic),
		version,
		headerCode,
		byte(f.fieldCount),
	))
	if err != nil {
		return err
	}

	_, err = w.w.Write(f.data)
	return err
}

func (w *Writer) WriteFrame(f *Fields, frameData []byte) error {
	// Frame header
	_, err := w.w.Write([]byte{frameCode, byte(f.fieldCount)})
	if err != nil {
		return err
	}

	// Frame fields
	_, err = w.w.Write(f.data)
	if err != nil {
		return err
	}

	// Frame thermal data
	_, err = w.w.Write(frameData)
	return err
}

func (w *Writer) Close() error {
	return w.w.Close()
}

func NewFields() *Fields {
	return &Fields{
		data: make([]byte, 0, 128),
	}
}

type Fields struct {
	data       []byte
	fieldCount uint8
}

func (f *Fields) Uint8(code byte, v uint8) {
	f.data = append(f.data, byte(1), code, byte(v))
	f.fieldCount++
}

func (f *Fields) Uint16(code byte, v uint16) {
	b := []byte{2, code, 0, 0}
	binary.LittleEndian.PutUint16(b[2:], v)
	f.data = append(f.data, b...)
	f.fieldCount++
}

func (f *Fields) Uint32(code byte, v uint32) {
	b := []byte{4, code, 0, 0, 0, 0}
	binary.LittleEndian.PutUint32(b[2:], v)
	f.data = append(f.data, b...)
	f.fieldCount++
}

func (f *Fields) Uint64(code byte, v uint64) {
	b := []byte{8, code, 0, 0, 0, 0, 0, 0, 0, 0}
	binary.LittleEndian.PutUint64(b[2:], v)
	f.data = append(f.data, b...)
	f.fieldCount++
}

func (f *Fields) Timestamp(code byte, t time.Time) {
	f.Uint64(code, uint64(t.UnixNano()*1000))
}
