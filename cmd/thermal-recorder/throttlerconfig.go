package main

type ThrottlerConfig struct {
	ApplyThrottling bool   `yaml:"apply-throttling"`
	ThrottleAfter   uint16 `yaml:"throttle-after-secs"`
	SparseAfter     uint16 `yaml:"sparse-after-secs"`
	SparseLength    uint16 `yaml:"sparse-length-secs"`
}

func DefaultThrottlerConfig() ThrottlerConfig {
	return ThrottlerConfig{
		ApplyThrottling: true,
		SparseAfter:     3600,
		SparseLength:    30,
		ThrottleAfter:   600,
	}
}
