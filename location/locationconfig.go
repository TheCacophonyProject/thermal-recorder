package location

import "errors"

const (
	defaultConfig = "/etc/cacophony/location.yaml"
	maxLatitude   = 90
	maxLongitude  = 180
)

type LocationConfig struct {
	Latitude  float64 `yaml:"latitude"`
	Longitude float64 `yaml:"longitude"`
}

func DefaultConfig() string {
	return defaultConfig
}

func DefaultLocationConfig() LocationConfig {
	return LocationConfig{
		Latitude:  -43.5321,
		Longitude: 172.6362,
	}
}
func (conf *LocationConfig) Validate() error {
	if &conf.Latitude == nil {
		return errors.New("Latitude cannot be nil")
	}
	if conf.Latitude < -maxLatitude || conf.Latitude > maxLatitude {
		return errors.New("Latitude outisde of normal range")
	}

	if &conf.Longitude == nil {
		return errors.New("Longitude cannot be nil")
	}
	if conf.Longitude < -maxLongitude || conf.Longitude > maxLongitude {
		return errors.New("Longitude outisde of normal range")
	}
	return nil
}
