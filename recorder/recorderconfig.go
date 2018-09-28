package recorder

import (
	"errors"

	"github.com/TheCacophonyProject/window"
)

type RecorderConfig struct {
	MinSecs     int              `yaml:"min-secs"`
	MaxSecs     int              `yaml:"max-secs"`
	PreviewSecs int              `yaml:"preview-secs"`
	WindowStart window.TimeOfDay `yaml:"window-start"`
	WindowEnd   window.TimeOfDay `yaml:"window-end"`
}

func DefaultRecorderConfig() RecorderConfig {
	return RecorderConfig{
		MinSecs:     10,
		MaxSecs:     600,
		PreviewSecs: 3,
	}
}

func (conf *RecorderConfig) Validate() error {
	if conf.MaxSecs < conf.MinSecs {
		return errors.New("max-secs should be larger than min-secs")
	}
	if conf.WindowStart.IsZero() && !conf.WindowEnd.IsZero() {
		return errors.New("window-end is set but window-start isn't")
	}
	if !conf.WindowStart.IsZero() && conf.WindowEnd.IsZero() {
		return errors.New("window-start is set but window-end isn't")
	}
	return nil
}
