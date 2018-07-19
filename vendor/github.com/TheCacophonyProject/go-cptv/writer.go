package cptv

import (
	"io"
	"time"

	"github.com/TheCacophonyProject/lepton3"
)

func NewWriter(w io.Writer) *Writer {
	return &Writer{
		bldr: NewBuilder(w),
		comp: NewCompressor(lepton3.FrameCols, lepton3.FrameRows),
	}
}

// Writer uses a Builder and Compressor to create CPTV files.
type Writer struct {
	bldr *Builder
	comp *Compressor
	t0   time.Time
}

func (w *Writer) WriteHeader(deviceName string) error {
	w.t0 = time.Now()
	fields := NewFieldWriter()
	fields.Timestamp(Timestamp, w.t0)
	fields.Uint32(XResolution, lepton3.FrameCols)
	fields.Uint32(YResolution, lepton3.FrameRows)
	fields.Uint8(Compression, 1)

	// Optional device name field
	if len(deviceName) > 0 {
		err := fields.String(DeviceName, deviceName)
		if err != nil {
			return err
		}
	}

	return w.bldr.WriteHeader(fields)
}

func (w *Writer) WriteFrame(prevFrame, frame *lepton3.Frame) error {
	dt := uint64(time.Since(w.t0))
	bitWidth, compFrame := w.comp.Next(prevFrame, frame)
	fields := NewFieldWriter()
	fields.Uint32(Offset, uint32(dt/1000))
	fields.Uint8(BitWidth, uint8(bitWidth))
	fields.Uint32(FrameSize, uint32(len(compFrame)))
	return w.bldr.WriteFrame(fields, compFrame)
}

func (w *Writer) Close() error {
	return w.bldr.Close()
}
