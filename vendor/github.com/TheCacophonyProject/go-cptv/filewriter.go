package cptv

import (
	"bufio"
	"os"
)

func NewFileWriter(filename string) (*FileWriter, error) {
	f, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	bw := bufio.NewWriter(f)
	return &FileWriter{
		Writer: NewWriter(bw),
		bw:     bw,
		f:      f,
	}, nil
}

// FileWriter wraps a Writer and provides a convenient way of writing
// a CPTV stream to a disk file.
type FileWriter struct {
	*Writer
	bw *bufio.Writer
	f  *os.File
}

func (fw *FileWriter) Name() string {
	return fw.f.Name()
}

func (fw *FileWriter) Close() {
	fw.Writer.Close()
	fw.bw.Flush()
	fw.f.Close()
}
