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

type service struct{}

func startService() error {
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
func (s service) TakeSnapshot() *dbus.Error {
	err := newSnapshot()
	if err != nil {
		return &dbus.Error{
			Name: dbusName + ".StayOnForError",
			Body: []interface{}{err.Error()},
		}
	}
	return nil
}
