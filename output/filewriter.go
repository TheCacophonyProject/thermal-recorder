package output

import (
	"bufio"
	"os"
	"time"

	"github.com/TheCacophonyProject/lepton3"
)

func NewFileWriter(filename string) (*FileWriter, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	bw := bufio.NewWriter(f)
	w := NewWriter(bw)
	return &FileWriter{
		f:    f,
		bw:   bw,
		w:    w,
		comp: NewCompressor(lepton3.FrameCols, lepton3.FrameRows),
	}, nil
}

type FileWriter struct {
	f    *os.File
	bw   *bufio.Writer
	w    *Writer
	comp *Compressor
	t0   time.Time
}

func (fw *FileWriter) WriteHeader() error {
	fw.t0 = time.Now()
	fields := NewFields()
	fields.Timestamp(Timestamp, fw.t0)
	fields.Uint32(XResolution, lepton3.FrameCols)
	fields.Uint32(YResolution, lepton3.FrameRows)
	fields.Uint8(Compression, 1)
	return fw.w.WriteHeader(fields)
}

func (fw *FileWriter) WriteFrame(prevFrame, frame *lepton3.Frame) error {
	dt := uint64(time.Since(fw.t0))

	bitWidth, compFrame := fw.comp.Next(prevFrame, frame)

	fields := NewFields()
	fields.Uint32(Offset, uint32(dt/1000))
	fields.Uint8(BitWidth, uint8(bitWidth))
	fields.Uint32(FrameSize, uint32(len(compFrame)))
	return fw.w.WriteFrame(fields, compFrame)
}

func (fw *FileWriter) Name() string {
	return fw.f.Name()
}

func (fw *FileWriter) Close() {
	fw.w.Close()
	fw.bw.Flush()
	fw.f.Close()
}
