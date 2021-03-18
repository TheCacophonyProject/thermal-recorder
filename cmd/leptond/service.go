// thermal-recorder - record thermal video footage of warm moving objects
//  Copyright (C) 2020, The Cacophony Project
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
	"fmt"
	"sync"

	"github.com/TheCacophonyProject/lepton3"
	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

const (
	dbusName = "org.cacophony.leptond"
	dbusPath = "/org/cacophony/leptond"
)

var mu sync.Mutex

type leptondService struct {
	camera *lepton3.Lepton3
}

func startService() (*leptondService, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	reply, err := conn.RequestName(dbusName, dbus.NameFlagDoNotQueue)
	if err != nil {
		return nil, err
	}
	if reply != dbus.RequestNameReplyPrimaryOwner {
		return nil, errors.New("name already taken")
	}
	s := &leptondService{}
	conn.Export(s, dbusPath, dbusName)
	conn.Export(genIntrospectable(s), dbusPath, "org.freedesktop.DBus.Introspectable")
	return s, nil
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

func (s *leptondService) setCamera(camera *lepton3.Lepton3) {
	mu.Lock()
	defer mu.Unlock()
	s.camera = camera
}

func (s *leptondService) removeCamera() {
	mu.Lock()
	defer mu.Unlock()
	s.camera = nil
}

func (s leptondService) RunFFC() *dbus.Error {
	mu.Lock()
	defer mu.Unlock()
	if s.camera == nil {
		return makeDbusError("RunFFC", fmt.Errorf("no camera available"))
	}
	if err := s.camera.RunFFC(); err != nil {
		return makeDbusError("RunFFC", err)
	}
	return nil
}

func (s leptondService) SetAutoFFC(automatic bool) *dbus.Error {
	mu.Lock()
	defer mu.Unlock()
	if s.camera == nil {
		return makeDbusError("SetAutoFFC", fmt.Errorf("no camera available"))
	}
	if err := s.camera.SetAutoFFC(automatic); err != nil {
		return makeDbusError("SetAutoFFC", err)
	}
	return nil
}

func makeDbusError(name string, err error) *dbus.Error {
	return &dbus.Error{
		Name: dbusName + "." + name,
		Body: []interface{}{err.Error()},
	}
}
