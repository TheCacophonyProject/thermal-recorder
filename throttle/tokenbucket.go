package throttle

// TokenBucket represents a bucket you can add or remove tokens from.   It will always have
// between 0 and size tokens in it (inclusive)
type TokenBucket struct {
	tokens uint32
	size   uint32
}

func (bucket *TokenBucket) AddTokens(newTokens uint32) {
	bucket.tokens += newTokens
	if bucket.tokens > bucket.size {
		bucket.tokens = bucket.size
	}
}

func (bucket *TokenBucket) RemoveTokens(oldTokens uint32) {
	if bucket.tokens >= oldTokens {
		bucket.tokens -= oldTokens
	} else {
		bucket.tokens = 0
	}
}

func (bucket *TokenBucket) HasTokens(tokens uint32) bool {
	return bucket.tokens >= tokens
}

func (bucket *TokenBucket) Empty() {
	bucket.tokens = 0
}

func (bucket *TokenBucket) IsFull() bool {
	return bucket.HasTokens(bucket.size)
}
