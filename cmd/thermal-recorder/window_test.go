package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNoWindow(t *testing.T) {
	zero := time.Time{}
	w := NewWindow(zero, zero)
	assert.True(t, w.Active())
}

func TestSameStartEnd(t *testing.T) {
	// Treat this as "no window"
	now := time.Now()
	w := NewWindow(now, now)
	assert.True(t, w.Active())
}

func TestStartLessThanEnd(t *testing.T) {
	w := NewWindow(mkTime(9, 10), mkTime(17, 30))

	w.Now = mkNow(9, 9)
	assert.False(t, w.Active())

	w.Now = mkNow(9, 10)
	assert.True(t, w.Active())

	w.Now = mkNow(12, 0)
	assert.True(t, w.Active())

	w.Now = mkNow(17, 30)
	assert.True(t, w.Active())

	w.Now = mkNow(17, 31)
	assert.False(t, w.Active())
}

func TestStartGreaterThanEnd(t *testing.T) {
	// Window goes over midnight
	w := NewWindow(mkTime(22, 10), mkTime(9, 50))

	w.Now = mkNow(22, 9)
	assert.False(t, w.Active())

	w.Now = mkNow(22, 10)
	assert.True(t, w.Active())

	w.Now = mkNow(23, 59)
	assert.True(t, w.Active())

	w.Now = mkNow(0, 0)
	assert.True(t, w.Active())

	w.Now = mkNow(0, 1)
	assert.True(t, w.Active())

	w.Now = mkNow(2, 0)
	assert.True(t, w.Active())

	w.Now = mkNow(9, 49)
	assert.True(t, w.Active())

	w.Now = mkNow(9, 50)
	assert.True(t, w.Active())

	w.Now = mkNow(9, 51)
	assert.False(t, w.Active())
}

func mkTime(hour, minute int) time.Time {
	return time.Date(1, 1, 1, hour, minute, 0, 0, time.UTC)
}

func mkNow(hour, minute int) func() time.Time {
	return func() time.Time {
		return time.Date(2017, 1, 2, hour, minute, 0, 0, time.UTC)
	}
}
