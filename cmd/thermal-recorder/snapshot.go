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
	"github.com/TheCacophonyProject/go-cptv/cptvframe"
	"log"
	"os"
	"path"
	"sync"
	"time"
)

const (
	snapshotName          = "still.png"
	allowedSnapshotPeriod = 500 * time.Millisecond
)

var (
	previousSnapshotID   = 0
	previousSnapshotTime time.Time
	mu                   sync.Mutex
)

func getHeaders() *headers.HeaderInfo {
	return header
}

func newSnapshot(dir string) (*cptvframe.Frame, error) {
	mu.Lock()
	defer mu.Unlock()

	if time.Since(previousSnapshotTime) < allowedSnapshotPeriod {
		return nil, nil
	}

	if processor == nil {
		return nil, errors.New("reading from camera has not started yet")
	}
	f := processor.GetRecentFrame()
	if f == nil {
		return nil, errors.New("no frames yet")
	}
	return f, nil
	// g16 := image.NewGray16(image.Rect(0, 0, lepton3.FrameCols, lepton3.FrameRows))
	// // Max and min are needed for normalization of the frame
	// var valMax uint16
	// var valMin uint16 = math.MaxUint16
	// var id int
	// for _, row := range f.Pix {
	// 	for _, val := range row {
	// 		id += int(val)
	// 		valMax = maxUint16(valMax, val)
	// 		valMin = minUint16(valMin, val)
	// 	}
	// }
	//
	// // Check if frame had already been processed
	// if id == previousSnapshotID {
	// 	return nil
	// }
	// previousSnapshotID = id
	//
	// var norm = math.MaxUint16 / (valMax - valMin)
	// for y, row := range f.Pix {
	// 	for x, val := range row {
	// 		g16.SetGray16(x, y, color.Gray16{Y: (val - valMin) * norm})
	// 	}
	// }
	//
	// out, err := os.Create(path.Join(dir, snapshotName))
	// if err != nil {
	// 	return err
	// }
	// defer out.Close()
	//
	// if err := png.Encode(out, g16); err != nil {
	// 	return err
	// }
	//
	// // the time will be changed only if the attempt is successful
	// previousSnapshotTime = time.Now()
	// return nil
}

func deleteSnapshot(dir string) {
	if err := os.Remove(path.Join(dir, snapshotName)); err != nil && !os.IsNotExist(err) {
		log.Printf("error with deleting snapshot image %s", err)
	}
}

func maxUint16(a, b uint16) uint16 {
	if a > b {
		return a
	}
	return b
}

func minUint16(a, b uint16) uint16 {
	if a < b {
		return a
	}
	return b
}
