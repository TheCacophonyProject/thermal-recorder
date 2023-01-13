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

package main

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"github.com/TheCacophonyProject/thermal-recorder/motion"
	"github.com/TheCacophonyProject/window"
)

const (
	allowedSnapshotPeriod = 500 * time.Millisecond
)

var (
	previousSnapshotID   = 0
	previousSnapshotTime time.Time
	mu                   sync.Mutex
)

func newSnapshot(lastFrame int) (*cptvframe.Frame, error) {
	mu.Lock()
	defer mu.Unlock()

	if time.Since(previousSnapshotTime) < allowedSnapshotPeriod {
		return nil, nil
	}
	if processor == nil {
		return nil, errors.New("reading from camera has not started yet")
	}
	if lastFrame >= 0 && uint32(lastFrame) == processor.CurrentFrame {
		return nil, errors.New("no new frames yet")
	}

	frameNum, f := processor.GetRecentFrame()
	if f == nil {
		return nil, errors.New("no frames yet")
	}
	if f.Status.FrameCount == 0 {
		f.Status.FrameCount = int(frameNum)
	}
	return f, nil
}

func newSnapshotRecording() error {
	mu.Lock()
	defer mu.Unlock()

	if processor == nil {
		return errors.New("reading from camera has not started yet")
	}

	processor.StartSnapshot = true
	return nil
}

func snapshotRecordingTriggers(processor *motion.MotionProcessor, window window.Window) {
	if window.NoWindow {
		log.Println("no recording window so will make snapshot every 12 hours")
		triggerTime := time.Now().Add(time.Minute)
		for {
			time.Sleep(time.Until(triggerTime))
			log.Println("making snapshot")
			processor.StartSnapshot = true
			triggerTime = triggerTime.Add(time.Hour * 12)
		}
	}

	// Make snapshot at start of window.
	sleepTime := time.Minute
	if !window.Active() {
		sleepTime = time.Until(window.NextStart()) + time.Minute
	}
	time.Sleep(sleepTime)
	log.Println("making start of window snapshot")
	processor.StartSnapshot = true

	// Make snapshot at end of window.
	time.Sleep(time.Until(window.NextEnd()) - 2*time.Minute)
	log.Println("making end of window snapshot")
	processor.StartSnapshot = true
}
