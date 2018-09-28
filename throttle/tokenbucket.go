package throttle

// TokenBucket represents a bucket you can add or remove tokens from.   It will always have
// between 0 and size tokens in it (inclusive)
type TokenBucket struct {
	tokens uint32
	size   uint32
}

// AddTokens Adds the specified number of tokens to the bucket (or as many as it can without
// overflowing the bucket)
func (bucket *TokenBucket) AddTokens(newTokens uint32) {
	bucket.tokens += newTokens
	if bucket.tokens > bucket.size {
		bucket.tokens = bucket.size
	}
}

// RemovesTokens Removes the specified number of tokens from the bucket, (or empties the bucket if it contains less
// than the specified number)
func (bucket *TokenBucket) RemoveTokens(oldTokens uint32) {
	if bucket.tokens >= oldTokens {
		bucket.tokens -= oldTokens
	} else {
		bucket.tokens = 0
	}
}

// HasTokens Returns true if the bucket has the specified number of tokens, else returns false.
func (bucket *TokenBucket) HasTokens(tokens uint32) bool {
	return bucket.tokens >= tokens
}

// Empty Empties the bucket
func (bucket *TokenBucket) Empty() {
	bucket.tokens = 0
}

// IsFull Returns true if the bucket is full, else returns false.
func (bucket *TokenBucket) IsFull() bool {
	return bucket.HasTokens(bucket.size)
}
