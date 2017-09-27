// Copyright 2017 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package lepton3

// ring manages a fixed byte slice, returning equal sized chunks of it
// with every call to next(). It is used to avoid memory allocation
// and garbage collection in frequently called code.
type ring struct {
	numChunks int
	chunkSize int
	ringSize  int
	offset    int
	buf       []byte
}

func newRing(numChunks, chunkSize int) *ring {
	ringSize := numChunks * chunkSize
	return &ring{
		numChunks: numChunks,
		chunkSize: chunkSize,
		ringSize:  ringSize,
		buf:       make([]byte, ringSize),
	}
}

func (r *ring) next() []byte {
	out := r.buf[r.offset : r.offset+r.chunkSize]
	r.offset += r.chunkSize
	if r.offset >= r.ringSize {
		r.offset = 0
	}
	return out
}
