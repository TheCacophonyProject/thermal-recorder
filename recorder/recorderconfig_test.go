package recorder

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v2"
)

func TestInvalidWindowStart(t *testing.T) {
	var config RecorderConfig
	err := yaml.UnmarshalStrict([]byte(`"window-start": 25:10`), &config)
	log.Printf("%v", config.WindowStart.Time)
	log.Printf("%v", err)
	log.Printf("%s", err)
	err2 := yaml.Unmarshal([]byte(`"min-secs": 25:10`), &config)
	log.Printf("%s", err2)

	assert.Equal(t, "invalid window-start", err)
}

// func TestInvalidWindowEnd(t *testing.T) {
// 	conf, err := ParseConfig([]byte("window-end: 25:10"), []byte(""))
// 	assert.Nil(t, conf)
// 	assert.EqualError(t, err, "invalid window-end")
// }

// func TestWindowEndWithoutStart(t *testing.T) {
// 	conf, err := ParseConfig([]byte("window-end: 09:10"), []byte(""))
// 	assert.Nil(t, conf)
// 	assert.EqualError(t, err, "window-end is set but window-start isn't")
// }

// func TestWindowStartWithoutEnd(t *testing.T) {
// 	conf, err := ParseConfig([]byte("window-start: 09:10"), []byte(""))
// 	assert.Nil(t, conf)
// 	assert.EqualError(t, err, "window-start is set but window-end isn't")
// }
