package output

// XXX docs!
// XXX tests!

import (
	"bytes"
	"encoding/binary"
	"io"
	"math"

	"github.com/TheCacophonyProject/lepton3"
)

type Compressor struct {
	cols, rows int
	frameDelta []int32
	adjDeltas  []int32
	outBuf     *bytes.Buffer
}

func NewCompressor(cols, rows int) *Compressor {
	elems := rows * cols
	outBuf := new(bytes.Buffer)
	outBuf.Grow(2 * elems) // 16 bits per element; worst case
	return &Compressor{
		rows:       rows,
		cols:       cols,
		frameDelta: make([]int32, elems),
		adjDeltas:  make([]int32, (elems)-1),
		outBuf:     outBuf,
	}
}

func (c *Compressor) Next(prev, curr *lepton3.Frame) (uint8, []byte) {
	// Generate the interframe delta.
	// The output is written in a "snaked" fashion to avoid
	// potentially greater deltas at the edges in the next stage.
	var i int
	for y := 0; y < c.rows; y++ {
		i = y * c.cols
		if y%2 == 1 {
			i += c.cols - 1
		}
		for x := 0; x < c.cols; x++ {
			c.frameDelta[i] = int32(curr[y][x]) - int32(prev[y][x])
			if y%2 == 0 {
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

// PackBits takes a slice of signed integers and packs them into an
// abitrary (smaller) bit width. The most significant bit is written
// out first.
func PackBits(width uint8, input []int32, w io.ByteWriter) {
	var bits uint32 // scratch buffer
	var nBits uint8 // number of bits in use in scratch
	for _, d := range input {
		bits |= twosComp(d, width) << (32 - width - nBits)
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
	widthMask := uint32((1 << width) - 1) // all 1's for the target width
	return uint32(-(v+1))&widthMask ^ widthMask
}

func numBits(x uint32) uint8 {
	return uint8(math.Floor(math.Log2(float64(x)) + 1))
}
