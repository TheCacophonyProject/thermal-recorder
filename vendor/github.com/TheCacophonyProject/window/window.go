// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package window

import (
	"time"
)

// New creates a Window instance which represents a recurring window
// between two times of day. If `start` is after `end` then the time
// window is assumed to cross over midnight. If `start` and `end` are
// the same then the window is always active.
func New(start, end time.Time) *Window {
	start = normaliseTime(start)
	end = normaliseTime(end)

	xMidnight := false
	if end.Before(start) {
		end = addDay(end)
		xMidnight = true
	}

	return &Window{
		Start:     start,
		End:       end,
		Now:       time.Now,
		xMidnight: xMidnight,
	}
}

// Window represents a recurring window between two times of day.
// The Now field can be use to override the time source (for testing).
type Window struct {
	Start     time.Time
	End       time.Time
	Now       func() time.Time
	xMidnight bool
}

// Active returns true if the time window is currently active.
func (w *Window) Active() bool {
	return w.Until() == time.Duration(0)
}

// Until returns the duration until the next time window starts.
func (w *Window) Until() time.Duration {
	if w.Start == w.End {
		return time.Duration(0)
	}

	now := w.nowTimeAfterStart()

	if w.End.After(now) {
		// During window.
		return time.Duration(0)
	}
	// After window.
	return addDay(w.Start).Sub(now)
}

func addDay(t time.Time) time.Time {
	return t.Add(24 * time.Hour)
}

// nowTimeAfterStart to make time calculations easier we choose a date time to represent now
// that is on or the next one after the start time
func (w *Window) nowTimeAfterStart() time.Time {
	now := normaliseTime(w.Now())
	if now.Before(w.Start) {
		now = addDay(now)
	}
	return now
}

func normaliseTime(t time.Time) time.Time {
	return time.Date(1, 1, 1, t.Hour(), t.Minute(), t.Second(), 0, time.UTC)
}

// UntilEnd returns the duration until the end of the time window.
func (w *Window) UntilEnd() time.Duration {
	if (w.Active()) {
		return w.End.Sub(w.nowTimeAfterStart())
	}
	return time.Duration(0)
}

// UntilNextInterval gets when the next interval starts.
// Only works when window is currently active.
func (w *Window) UntilNextInterval(interval time.Duration) time.Duration {
	if (w.Active()) {
		now := w.nowTimeAfterStart()
		elapsedTime := now.Sub(w.Start)
		nextInterval := w.Start.Add(elapsedTime.Truncate(interval) + interval)

		if (w.End.After(nextInterval)) {
			return nextInterval.Sub(now)
		}
	}

	return time.Duration(-1)
}
