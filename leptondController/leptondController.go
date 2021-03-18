package leptondController

import "github.com/godbus/dbus"

const (
	dbusPath   = "/org/cacophony/leptond"
	dbusDest   = "org.cacophony.leptond"
	methodBase = "org.cacophony.leptond"
)

func getDbusObj() (dbus.BusObject, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, err
	}
	obj := conn.Object(dbusDest, dbusPath)
	return obj, nil
}

func SetAutoFFC(automatic bool) error {
	obj, err := getDbusObj()
	if err != nil {
		return err
	}
	return obj.Call(methodBase+".SetAutoFFC", 0, automatic).Store()
}

func RunFFC() error {
	obj, err := getDbusObj()
	if err != nil {
		return err
	}
	return obj.Call(methodBase+".RunFFC", 0).Store()
}
