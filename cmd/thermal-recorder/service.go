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

	"github.com/godbus/dbus"
	"github.com/godbus/dbus/introspect"
)

const (
	dbusName = "org.cacophony.thermalrecorder"
	dbusPath = "/org/cacophony/thermalrecorder"
)

type service struct {
	dir string
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

	s := &service{
		dir: dir,
	}
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
func (s *service) TakeSnapshot() *dbus.Error {
	err := newSnapshot(s.dir, false)
	if err != nil {
		return &dbus.Error{
			Name: dbusName + ".TakeSnapshot",
			Body: []interface{}{err.Error()},
		}
	}
	return nil
}

// TakeRawSnapshot will save the next frame as a unnormalised still
func (s *service) TakeRawSnapshot() *dbus.Error {
	err := newSnapshot(s.dir, true)
	if err != nil {
		return &dbus.Error{
			Name: dbusName + ".TakeRawSnapshot",
			Body: []interface{}{err.Error()},
		}
	}
	return nil
}
