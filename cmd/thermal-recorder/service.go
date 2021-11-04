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
	"github.com/TheCacophonyProject/thermal-recorder/headers"

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

const (
	dbusName = "org.cacophony.thermalrecorder"
	dbusPath = "/org/cacophony/thermalrecorder"
)

type service struct {
}

func startService(dir string) error {
	conn, err := dbus.SystemBus()
	if err != nil {
		return err
	}
	reply, err := conn.RequestName(dbusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return errors.New("name already taken")
	}

	s := &service{}
	conn.Export(s, dbusPath, dbusName)
	conn.Export(genIntrospectable(s), dbusPath, "org.freedesktop.DBus.Introspectable")
	return nil
}

func genIntrospectable(v interface{}) introspect.Introspectable {
	node := &introspect.Node{
		Interfaces: []introspect.Interface{{
			Name:    dbusName,
			Methods: introspect.Methods(v),
		}},
	}
	return introspect.NewIntrospectable(node)
}

// TakeSnapshot will save the next frame as a still
func (s *service) TakeSnapshot(lastFrame int) (*cptvframe.Frame, *dbus.Error) {
	f, err := newSnapshot(lastFrame)
	if err != nil {
		return nil, &dbus.Error{
			Name: dbusName + ".TakeSnapshot",
			Body: []interface{}{err.Error()},
		}
	}
	return f, nil
}
func (s *service) CameraInfo() (map[string]interface{}, *dbus.Error) {

	if headerInfo == nil {
		return nil, &dbus.Error{
			Name: dbusName + ".NoHeaderInfo",
			Body: nil,
		}
	}
	camera_specs := map[string]interface{}{
		headers.XResolution: headerInfo.ResX(),
		headers.YResolution: headerInfo.ResY(),
		headers.FrameSize:   headerInfo.FrameSize(),
		headers.Model:       headerInfo.Model(),
		headers.Brand:       headerInfo.Brand(),
		headers.FPS:         headerInfo.FPS(),
		headers.Serial:      headerInfo.CameraSerial(),
		headers.Firmware:    headerInfo.Firmware(),
	}
	return camera_specs, nil
}
