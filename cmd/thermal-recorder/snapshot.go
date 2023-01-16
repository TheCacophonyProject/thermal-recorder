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
		log.Println("no motion processor so can't make snapshot")
		return errors.New("reading from camera has not started yet")
	}

	processor.StartSnapshot = true
	return nil
}

// snapshotRecordingTriggers will make a snapshot when in the recording window and at the end of the recording window.
func snapshotRecordingTriggers(window window.Window) {

	// Wait for motion processor to start
	for processor == nil {
		time.Sleep(time.Second)
	}

	if window.NoWindow {
		log.Println("no recording window so will make snapshot every 12 hours")
		triggerTime := time.Now().Add(time.Minute)
		for {
			time.Sleep(time.Until(triggerTime))
			_ = newSnapshotRecording()
			triggerTime = triggerTime.Add(time.Hour * 12)
		}
	}

	if window.Active() {
		// If camera just started give it a minute to warm  up.
		time.Sleep(time.Minute)
		log.Println("making power on snapshot")
	} else {
		// Wait for recording window to start
		time.Sleep(time.Until(window.NextStart()) + time.Minute)
		log.Println("making start of window snapshot")
	}
	_ = newSnapshotRecording()

	// Make snapshot at end of window.
	time.Sleep(time.Until(window.NextEnd()) - 2*time.Minute)
	log.Println("making end of window snapshot")
	_ = newSnapshotRecording()
}
