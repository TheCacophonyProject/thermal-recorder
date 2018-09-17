package main

type ThrottlerConfig struct {
	ThrottleAfter     uint16 `yaml:"throttle-after-secs"`
	OccasionalAfter   uint16 `yaml:"occ-after-secs"`
	OccassionalLength uint16 `yaml:"occ-length-secs"`
}

func DefaultThrottlerConfig() ThrottlerConfig {
	return ThrottlerConfig{
		OccasionalAfter:   3600,
		OccassionalLength: 30,
		ThrottleAfter:     600,
	}
}
