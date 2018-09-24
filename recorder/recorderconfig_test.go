package recorder

import (
	"testing"

	"github.com/TheCacophonyProject/window"

	"github.com/stretchr/testify/assert"
)

func TestWindowStartWithoutEndDoesntValidate(t *testing.T) {
	conf := RecorderConfig{
		WindowStart: *window.NewTimeOfDay("09:10"),
	}
	assert.EqualError(t, conf.Validate(), "window-start is set but window-end isn't")
}

func TestWindowEndWithoutStartDoesntValidate(t *testing.T) {
	conf := RecorderConfig{
		WindowEnd: *window.NewTimeOfDay("09:10"),
	}
	assert.EqualError(t, conf.Validate(), "window-end is set but window-start isn't")
}

func TestMinSecsGreaterThanMaxSecsDoesntValidate(t *testing.T) {
	conf := RecorderConfig{
		MinSecs: 5,
		MaxSecs: 2,
	}
	assert.EqualError(t, conf.Validate(), "max-secs should be larger than min-secs")
}
