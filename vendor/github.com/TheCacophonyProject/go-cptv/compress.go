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
	"bytes"
	"encoding/binary"
	"io"
	"math"

	"github.com/TheCacophonyProject/lepton3"
)

// NewCompressor creates a new Compressor.
func NewCompressor() *Compressor {
	elems := lepton3.FrameRows * lepton3.FrameCols
	outBuf := new(bytes.Buffer)
	outBuf.Grow(2 * elems) // 16 bits per element; worst case
	return &Compressor{
		rows:       lepton3.FrameRows,
		cols:       lepton3.FrameCols,
		frameDelta: make([]int32, elems),
		adjDeltas:  make([]int32, elems-1),
		outBuf:     outBuf,
		prevFrame:  new(lepton3.Frame),
	}
}

// Compressor generates a compressed representation of successive
// lepton3 Frames, returning CPTV frames.
type Compressor struct {
	cols, rows int
	frameDelta []int32
	adjDeltas  []int32
	outBuf     *bytes.Buffer
	prevFrame  *lepton3.Frame
}

// Next takes the next lepton3.Frame in a recording and converts it to
// a compressed stream of bytes. The bit width used for packing is
// also returned (this is required for unpacking).
//
// IMPORTANT: The returned byte slice is reused and therefore is only
// valid until the next call to Next.
func (c *Compressor) Next(curr *lepton3.Frame) (uint8, []byte) {
	// Generate the interframe delta.
	// The output is written in a "snaked" fashion to avoid
	// potentially greater deltas at the edges in the next stage.
	var i int
	for y := 0; y < c.rows; y++ {
		i = y * c.cols
		if y&1 == 1 {
			i += c.cols - 1
		}
		for x := 0; x < c.cols; x++ {
			c.frameDelta[i] = int32(curr.Pix[y][x]) - int32(c.prevFrame.Pix[y][x])
			// Now that prevFrame[y][x] has been used, copy the value
			// for the current frame in for the next call to Next().
			// TODO: it might be fast to copy() rows separately.
			c.prevFrame.Pix[y][x] = curr.Pix[y][x]
			if y&1 == 0 {
				i++
			} else {
				i--
			}
		}
	}

	// Now generate the adjacent "delta of deltas".
	var maxD uint32
	for i := 0; i < len(c.frameDelta)-1; i++ {
		d := c.frameDelta[i+1] - c.frameDelta[i]
		c.adjDeltas[i] = d
		if absD := abs(d); absD > maxD {
			maxD = absD
		}
	}

	// How many bits required to store the largest delta?
	width := numBits(maxD) + 1 // add 1 to allow for sign bit

	// Write out the starting frame delta value (required for reconstruction)
	c.outBuf.Reset()
	binary.Write(c.outBuf, binary.LittleEndian, c.frameDelta[0])

	// Pack the deltas according to the bit width determined
	PackBits(width, c.adjDeltas, c.outBuf)
	return width, c.outBuf.Bytes()
}

// NewDecompressor creates a new Decompressor.
func NewDecompressor() *Decompressor {
	return &Decompressor{
		rows:       lepton3.FrameRows,
		cols:       lepton3.FrameCols,
		pixelCount: lepton3.FrameRows * lepton3.FrameCols,
		prevFrame:  new(lepton3.Frame),
	}
}

// Decompressor is used to decompress successive CPTV frames. See the
// Next() method.
type Decompressor struct {
	cols, rows, pixelCount int
	prevFrame              *lepton3.Frame
	deltas                 [lepton3.FrameRows][lepton3.FrameCols]int32
}

// ByteReaderReader combines io.Reader and io.ByteReader.
type ByteReaderReader interface {
	io.Reader
	io.ByteReader
}

// Next reads of stream of bytes as a ByteReaderReader and
// decompresses them using the bit width provided into the
// lepton3.Frame provided.
func (d *Decompressor) Next(bitWidth uint8, compressed ByteReaderReader, out *lepton3.Frame) error {
	var v int32
	err := binary.Read(compressed, binary.LittleEndian, &v)
	if err != nil {
		return err
	}

	unpacker := NewBitUnpacker(bitWidth, compressed)
	d.deltas[0][0] = v
	for i := 1; i < d.pixelCount; i++ {
		y := i / lepton3.FrameCols
		x := i % lepton3.FrameCols
		// Deltas are "snaked" so work backwards through every second row.
		if y&1 == 1 {
			x = lepton3.FrameCols - x - 1
		}

		dv, err := unpacker.Next()
		if err != nil {
			return err
		}
		v += dv
		d.deltas[y][x] = v
	}

	// Add to delta frame to previous frame.
	for y := 0; y < lepton3.FrameRows; y++ {
		for x := 0; x < lepton3.FrameCols; x++ {
			out.Pix[y][x] = uint16(int32(d.prevFrame.Pix[y][x]) + d.deltas[y][x])
			// Now that prevFrame[y][x] has been used, copy the new
			// value in for the next call to Next() to use.
			// TODO: it might be fast to copy() rows separately.
			d.prevFrame.Pix[y][x] = out.Pix[y][x]
		}
	}
	return nil
}

// PackBits takes a slice of signed integers and packs them into an
// abitrary (smaller) bit width. The most significant bit is written
// out first.
func PackBits(width uint8, input []int32, w io.ByteWriter) {
	var bits uint32 // scratch buffer
	var nBits uint8 // number of bits in use in scratch
	for _, d := range input {
		bits |= uint32(twosComp(d, width) << (32 - width - nBits))
		nBits += width
		for nBits >= 8 {
			w.WriteByte(uint8(bits >> 24))
			bits <<= 8
			nBits -= 8
		}
	}
	if nBits > 0 {
		w.WriteByte(uint8(bits >> 24))
	}
}

// NewBitUnpacker creates a new BitUnpacker. Integers will be
// extracted from the ByteReader and are expected to be packed at the
// bit width specified.
func NewBitUnpacker(width uint8, r io.ByteReader) *BitUnpacker {
	return &BitUnpacker{
		bitw: width,
		r:    r,
	}
}

// BitUnpacker extracts signed integers, packed at some bit width,
// from a bitstream.
type BitUnpacker struct {
	r     io.ByteReader
	bitw  uint8
	bits  uint32
	nbits uint8
}

// Next returns the next signed integer from the bitstream.
func (u *BitUnpacker) Next() (int32, error) {
	for u.nbits < u.bitw {
		b, err := u.r.ReadByte()
		if err != nil {
			return 0, err
		}
		u.bits |= uint32(b) << uint8(24-u.nbits)
		u.nbits += 8
	}

	out := twosUncomp(u.bits>>(32-u.bitw), u.bitw)
	u.bits = u.bits << u.bitw
	u.nbits -= u.bitw
	return out, nil
}

func abs(x int32) uint32 {
	if x < 0 {
		return uint32(-x)
	}
	return uint32(x)
}

func twosComp(v int32, width uint8) uint32 {
	if v >= 0 {
		return uint32(v)
	}
	return (^uint32(-v) + 1) & uint32((1<<width)-1)
}

func twosUncomp(v uint32, width uint8) int32 {
	if v&(1<<(width-1)) == 0 {
		return int32(v) // positive
	}
	return -int32((^v + 1) & uint32((1<<width)-1))
}

func numBits(x uint32) uint8 {
	return uint8(math.Floor(math.Log2(float64(x)) + 1))
}
