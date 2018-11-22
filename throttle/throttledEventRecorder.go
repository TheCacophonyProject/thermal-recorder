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

import (
	"encoding/json"
	"log"
	"time"

	"github.com/godbus/dbus"
)

// uses the event api to record that video was throttled at a particular time.
type ThrottledEventRecorder struct {
}

func (er ThrottledEventRecorder) WhenThrottled() {
	ts := time.Now()
	eventDetails := map[string]interface{}{
		"description": map[string]interface{}{
			"type": "throttle",
		},
	}
	detailsJSON, err := json.Marshal(&eventDetails)
	if err != nil {
		log.Printf("Could not record throttle event: %s", err)
		return
	}

	conn, err := dbus.SystemBus()
	if err != nil {
		log.Printf("Could not record throttle event: %s", err)
		return
	}

	obj := conn.Object("org.cacophony.Events", "/org/cacophony/Events")
	call := obj.Call("org.cacophony.Events.Queue", 0, detailsJSON, ts.UnixNano())
	if call.Err != nil {
		log.Printf("Could not record throttle event: %s", call.Err)
		return
	}
}
