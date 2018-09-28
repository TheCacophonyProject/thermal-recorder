package throttle

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestBucket struct {
	TokenBucket
}

func (b *TestBucket) hasExactly(tokens float64) bool {
	return b.HasTokens(tokens) && !b.HasTokens(tokens+1)
}

func TestTokenBucketCanAddTokens(t *testing.T) {
	bucket := TestBucket{TokenBucket{size: 5}}

	bucket.AddTokens(4)
	assert.True(t, bucket.hasExactly(4))

	bucket.AddTokens(0)
	assert.True(t, bucket.hasExactly(4))
	assert.False(t, bucket.IsFull())

	// check cannot over-flow bucket
	bucket.AddTokens(3)
	assert.True(t, bucket.hasExactly(5))
	assert.True(t, bucket.IsFull())
}

func TestTokenBucketCanRemoveTokens(t *testing.T) {
	bucket := TestBucket{TokenBucket{size: 5}}

	bucket.AddTokens(4)

	bucket.RemoveTokens(3)
	assert.True(t, bucket.hasExactly(1))

	// test removing too many tokens
	bucket.RemoveTokens(3)
	assert.True(t, bucket.hasExactly(0))
}

func TestTokenBucketEmptyRemovesAllTokensFromBucket(t *testing.T) {
	bucket := TestBucket{TokenBucket{size: 5}}

	bucket.AddTokens(4)
	bucket.Empty()
	assert.True(t, bucket.hasExactly(0))
}
