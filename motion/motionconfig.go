package motion

import "errors"

type MotionConfig struct {
	TempThresh        uint16 `yaml:"temp-thresh"`
	DeltaThresh       uint16 `yaml:"delta-thresh"`
	CountThresh       int    `yaml:"count-thresh"`
	NonzeroMaxPercent int    `yaml:"nonzero-max-percent"`
	FrameCompareGap   int    `yaml:"frame-compare-gap"`
	UseOneDiffOnly    bool   `yaml:"one-diff-only"`
	TriggerFrames     int    `yaml:"trigger-frames"`
	WarmerOnly        bool   `yaml:"warmer-only"`
	Verbose           bool   `yaml:"verbose"`
}

func DefaultMotionConfig() MotionConfig {
	return MotionConfig{
		TempThresh:        2900,
		DeltaThresh:       50,
		CountThresh:       3,
		NonzeroMaxPercent: 50,
		FrameCompareGap:   45,
		Verbose:           false,
		TriggerFrames:     2,
		UseOneDiffOnly:    true,
		WarmerOnly:        true,
	}
}

func (conf *MotionConfig) Validate() error {
	if conf.NonzeroMaxPercent < 1 || conf.NonzeroMaxPercent > 100 {
		return errors.New("nonzero-max-percent should be in range 1 - 100")
	}
	return nil
}
