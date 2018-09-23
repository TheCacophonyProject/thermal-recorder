// Copyright 2018 The Cacophony Project. All rights reserved.
// Use of this source code is governed by the Apache License Version 2.0;
// see the LICENSE file for further details.

package window_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/TheCacophonyProject/window"
)

func TestNoWindow(t *testing.T) {
	zero := time.Time{}
	w := window.New(zero, zero)
	assert.True(t, w.Active())
}

func TestSameStartEnd(t *testing.T) {
	// Treat this as "no window"
	now := time.Now()
	w := window.New(now, now)

	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
}

func TestStartLessThanEnd(t *testing.T) {
	w := window.New(mkTime(9, 10), mkTime(17, 30))
	interval := time.Duration(30 * time.Minute)

	w.Now = mkNow(9, 9)
	assert.False(t, w.Active())
	assert.Equal(t, time.Minute, w.Until())
	assert.Equal(t, time.Duration(-1), w.UntilNextInterval(interval))
	assert.Equal(t, time.Duration(0), w.UntilEnd())

	w.Now = mkNow(9, 10)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Duration(30*time.Minute), w.UntilNextInterval(interval))

	w.Now = mkNow(12, 0)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Duration(10*time.Minute), w.UntilNextInterval(interval))
	assert.Equal(t, time.Duration((5*60+30)*time.Minute), w.UntilEnd())

	w.Now = mkNow(17, 29)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Duration(-1), w.UntilNextInterval(interval))
	assert.Equal(t, time.Minute, w.UntilEnd())

	w.Now = mkNow(17, 30)
	assert.False(t, w.Active())
	assert.Equal(t, 940*time.Minute, w.Until())
	assert.Equal(t, time.Duration(-1), w.UntilNextInterval(interval))
	assert.Equal(t, time.Duration(0), w.UntilEnd())
}

func TestStartGreaterThanEnd(t *testing.T) {
	// Window goes over midnight
	w := window.New(mkTime(22, 10), mkTime(9, 50))
	interval := time.Duration(30 * time.Minute)

	w.Now = mkNow(22, 9)
	assert.False(t, w.Active())
	assert.Equal(t, time.Minute, w.Until())
	assert.Equal(t, time.Duration(-1), w.UntilNextInterval(interval))

	w.Now = mkNow(22, 10)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())

	w.Now = mkNow(23, 59)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Duration(11*time.Minute), w.UntilNextInterval(interval))
	assert.Equal(t, time.Duration((9*60+51)*time.Minute), w.UntilEnd())

	w.Now = mkNow(0, 0)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Duration(10*time.Minute), w.UntilNextInterval(interval))

	w.Now = mkNow(0, 1)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Duration(9*time.Minute), w.UntilNextInterval(interval))

	w.Now = mkNow(2, 0)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())

	w.Now = mkNow(9, 49)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Duration(time.Minute), w.UntilEnd())

	w.Now = mkNow(9, 49)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Duration(-1), w.UntilNextInterval(interval))

	w.Now = mkNow(9, 50)
	assert.False(t, w.Active())
	assert.Equal(t, 740*time.Minute, w.Until())
	assert.Equal(t, time.Duration(-1), w.UntilNextInterval(interval))
}

func TestMorningToMorning(t *testing.T) {
	// Window not active just between 10am and 11am each day.
	w := window.New(mkTime(11, 0), mkTime(10, 0))

	w.Now = mkNow(9, 59)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, time.Minute, w.UntilEnd())

	w.Now = mkNow(10, 0)
	assert.False(t, w.Active())
	assert.Equal(t, time.Hour, w.Until())

	w.Now = mkNow(10, 59)
	assert.False(t, w.Active())
	assert.Equal(t, time.Minute, w.Until())

	w.Now = mkNow(11, 0)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, 23*time.Hour, w.UntilEnd())

	w.Now = mkNow(18, 0)
	assert.True(t, w.Active())
	assert.Equal(t, time.Duration(0), w.Until())
	assert.Equal(t, 16*time.Hour, w.UntilEnd())
}

func mkTime(hour, minute int) time.Time {
	return time.Date(1, 1, 1, hour, minute, 0, 0, time.UTC)
}

func mkNow(hour, minute int) func() time.Time {
	return func() time.Time {
		return time.Date(2017, 1, 2, hour, minute, 0, 0, time.UTC)
	}
}
