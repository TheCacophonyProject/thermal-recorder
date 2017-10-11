package cptv

import (
	"io"
)

// nReader wraps an io.Reader providing some convenience methods for
// reading out specific lengths of data.
type nReader struct {
	io.Reader
}

func (r *nReader) ReadByteInt() (int, error) {
	b, err := r.ReadByte()
	if err != nil {
		return 0, err
	}
	return int(b), nil
}

func (r *nReader) ReadByte() (byte, error) {
	buf, err := r.ReadN(1)
	if err != nil {
		return 0, err
	}
	return buf[0], nil
}

func (r *nReader) ReadN(n int) ([]byte, error) {
	buf := make([]byte, n)
	i := 0
	for i < n {
		sz, err := r.Read(buf[i:])
		if err != nil {
			return nil, err
		}
		i += sz
	}
	return buf, nil
}
