package cptv

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTwosComp(t *testing.T) {
	tests := []struct {
		input    int32
		width    uint8
		expected uint32
	}{
		// Width 8
		{0, 8, 0},
		{1, 8, 1},
		{-1, 8, 255},
		{15, 8, 15},
		{-15, 8, 241},
		{127, 8, 127},
		{-127, 8, 129},
		{-128, 8, 128},

		{-12, 9, 500},

		// Width 5
		{0, 5, 0},
		{1, 5, 1},
		{-1, 5, 31},
		{15, 5, 15},
		{-15, 5, 17},
		{-16, 5, 16},

		// Width 14
		{0, 14, 0},
		{1, 14, 1},
		{-1, 14, 16383},
		{15, 14, 15},
		{-15, 14, 16369},
		{8191, 14, 8191},
		{-8192, 14, 8192},
	}

	for _, x := range tests {
		assert.Equal(
			t,
			x.expected,
			twosComp(x.input, x.width),
			"%d (width=%d)", x.input, x.width,
		)
	}
}
