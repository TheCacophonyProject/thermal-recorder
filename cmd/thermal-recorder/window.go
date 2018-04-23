package main

import (
	"time"
)

func NewWindow(start, end time.Time) *Window {
	start = normaliseTime(start)
	end = normaliseTime(end)

	xMidnight := false
	if end.Before(start) {
		end = end.Add(24 * time.Hour)
		xMidnight = true
	}

	return &Window{
		Start:     start,
		End:       end,
		Now:       time.Now,
		xMidnight: xMidnight,
	}
}

type Window struct {
	Start     time.Time
	End       time.Time
	Now       func() time.Time
	xMidnight bool
}

func (w *Window) Active() bool {
	if w.Start == w.End {
		return true
	}

	now := normaliseTime(w.Now())
	if w.xMidnight && now.Before(w.Start) {
		now = now.Add(24 * time.Hour)
	}
	return !now.Before(w.Start) && !now.After(w.End)
}

func normaliseTime(t time.Time) time.Time {
	return time.Date(1, 1, 1, t.Hour(), t.Minute(), 0, 0, time.UTC)
}
