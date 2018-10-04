// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2018, The Cacophony Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package throttle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestBucket struct {
	TokenBucket
}

func (b *TestBucket) hasApprox(tokens float64) bool {
	return b.HasTokens(tokens) && !b.HasTokens(tokens+1)
}

func TestTokenBucketCanAddTokens(t *testing.T) {
	bucket := TestBucket{TokenBucket{size: 5}}

	bucket.AddTokens(4)
	assert.True(t, bucket.hasApprox(4))

	bucket.AddTokens(0)
	assert.True(t, bucket.hasApprox(4))
	assert.False(t, bucket.IsFull())

	// check cannot over-flow bucket
	bucket.AddTokens(3)
	assert.True(t, bucket.hasApprox(5))
	assert.True(t, bucket.IsFull())
}

func TestTokenBucketCanRemoveTokens(t *testing.T) {
	bucket := TestBucket{TokenBucket{size: 5}}

	bucket.AddTokens(4)

	bucket.RemoveTokens(3)
	assert.True(t, bucket.hasApprox(1))

	// test removing too many tokens
	bucket.RemoveTokens(3)
	assert.True(t, bucket.hasApprox(0))
}

func TestTokenBucketEmptyRemovesAllTokensFromBucket(t *testing.T) {
	bucket := TestBucket{TokenBucket{size: 5}}

	bucket.AddTokens(4)
	bucket.Empty()
	assert.True(t, bucket.hasApprox(0))
}
