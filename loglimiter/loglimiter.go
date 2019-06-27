// thermal-recorder - record thermal video footage of warm moving objects
// Copyright (C) 2019, The Cacophony Project
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

package loglimiter

import (
	"fmt"
	"log"
	"time"
)

// New returns a new LogLimiter with the configured minimum log interval.
func New(interval time.Duration) *LogLimiter {
	return &LogLimiter{
		interval: interval,
		nowFunc:  time.Now,
	}
}

// LogLimiter will suppress log messages if the same log message is
// seen within some time interval.
type LogLimiter struct {
	interval      time.Duration
	nowFunc       func() time.Time
	previousEntry string
	previousTime  time.Time
}

func (limiter *LogLimiter) Printf(format string, v ...interface{}) {
	limiter.Print(fmt.Sprintf(format, v...))
}

func (limiter *LogLimiter) Print(s string) {
	now := limiter.nowFunc()
	if now.Sub(limiter.previousTime) < limiter.interval && s == limiter.previousEntry {
		return
	}

	log.Print(s)
	limiter.previousTime = now
	limiter.previousEntry = s
}
