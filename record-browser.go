package avahi

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
)

// A RecordBrowser is a browser for mDNS records
type RecordBrowser struct {
	object        dbus.BusObject
	AddChannel    chan Record
	RemoveChannel chan Record
	closeCh       chan struct{}
}

// RecordBrowserNew creates a new mDNS record browser
func RecordBrowserNew(conn *dbus.Conn, path dbus.ObjectPath) (*RecordBrowser, error) {
	c := new(RecordBrowser)

	c.object = conn.Object("org.freedesktop.Avahi", path)
	c.AddChannel = make(chan Record)
	c.RemoveChannel = make(chan Record)

	return c, nil
}

func (c *RecordBrowser) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.RecordBrowser", method)
}

func (c *RecordBrowser) free() {
	close(c.closeCh)
	close(c.AddChannel)
	close(c.RemoveChannel)
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *RecordBrowser) getObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *RecordBrowser) dispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("ItemNew") || signal.Name == c.interfaceForMember("ItemRemove") {
		var record Record
		err := dbus.Store(signal.Body, &record.Interface, &record.Protocol, &record.Name,
			&record.Class, &record.Type, &record.Rdata, &record.Flags)
		if err != nil {
			return err
		}

		if signal.Name == c.interfaceForMember("ItemNew") {
			select {
			case c.AddChannel <- record:
			case <-c.closeCh:
			}
		} else {
			select {
			case c.RemoveChannel <- record:
			case <-c.closeCh:
			}
		}
	}

	return nil
}
