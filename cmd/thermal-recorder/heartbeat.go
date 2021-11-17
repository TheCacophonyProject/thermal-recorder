package main

import (
	"github.com/TheCacophonyProject/window"
	"log"
	"time"
)

const interval = 3 * time.Hour

type HeartBeat struct {
	window    window.Window
	nextEvent time.Time
	start     time.Time
	end       time.Time
	constant  bool
}

func NewHeartBeat(window window.Window, constant bool) *HeartBeat {
	var nextStart time.Time
	var nextEnd time.Time
	if window.NoWindow {
		constant = true
	} else {
		nextStart = window.NextStart()
		nextEnd = window.NextEnd()
	}
	h := &HeartBeat{window, time.Now(), nextStart, nextEnd, constant}
	return h
}

func (h *HeartBeat) Check() error {
	if time.Now().After(h.nextEvent) {
		h.nextEvent = h.nextEvent.Add(interval)
		if !h.constant {
			if h.start.After(time.Now()) && h.nextEvent.After(h.start) {
				// always trigger on start if can
				h.nextEvent = h.start
			} else if h.nextEvent.After(h.end) {
				h.nextEvent = h.end
				// always trigger on end
			} else if h.nextEvent.After(h.end.Add(-1 * time.Hour)) {
				h.nextEvent = h.end.Add(-1 * time.Hour)
				// 1 hour before rec window ends
			}
			if h.window.NextEnd() != h.end {
				// New Window
				h.start = h.window.NextStart()
				h.end = h.window.NextEnd()
			}
		}
		h.sendEvent()
	}

	return nil
}

func (h *HeartBeat) sendEvent() error {
	// send api call with happy tunil h.nextEvent
	// send event
	log.Printf("Send heartbeat next event %v window %v - %v", h.nextEvent, h.start, h.end)
	return nil
}
