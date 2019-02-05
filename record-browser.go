package avahi

import (
	"fmt"
	"github.com/godbus/dbus"
)

type RecordBrowser struct {
	object        dbus.BusObject
	AddChannel    chan Record
	RemoveChannel chan Record
}

func RecordBrowserNew(conn *dbus.Conn, path dbus.ObjectPath) (*RecordBrowser, error) {
	c := new(RecordBrowser)

	c.object = conn.Object("org.freedesktop.Avahi.RecordBrowser", path)
	c.AddChannel = make(chan Record, 10)
	c.RemoveChannel = make(chan Record, 10)

	return c, nil
}

func (c *RecordBrowser) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.RecordBrowser", method)
}

func (c *RecordBrowser) free() {
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *RecordBrowser) GetObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *RecordBrowser) DispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("ItemNew") || signal.Name == c.interfaceForMember("ItemRemove") {
		var record Record
		err := dbus.Store(signal.Body, &record.Interface, &record.Protocol, &record.Name,
			&record.Class, &record.Type, &record.Rdata, &record.Flags)
		if err != nil {
			return err
		}

		if signal.Name == c.interfaceForMember("ItemNew") {
			c.AddChannel <- record
		} else {
			c.RemoveChannel <- record
		}
	}

	return nil
}
