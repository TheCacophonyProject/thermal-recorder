// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2018, The Cacophony Project
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program. If not, see <http://www.gnu.org/licenses/>.

package throttle

import "time"

type ThrottlerConfig struct {
	ApplyThrottling bool          `yaml:"apply-throttling"`
	BucketSize      time.Duration `yaml:"bucket-size"`
	MinRefill       time.Duration `yaml:"min-refill"`
}

func DefaultThrottlerConfig() ThrottlerConfig {
	return ThrottlerConfig{
		ApplyThrottling: true,
		BucketSize:      10 * time.Minute,
		MinRefill:       10 * time.Minute,
	}
}
