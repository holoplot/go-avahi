package avahi

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
)

// A ServiceBrowser browses for mDNS services
type ServiceBrowser struct {
	object        dbus.BusObject
	AddChannel    chan Service
	RemoveChannel chan Service
	closeCh       chan struct{}
}

// ServiceBrowserNew creates a new browser for mDNS records
func ServiceBrowserNew(conn *dbus.Conn, path dbus.ObjectPath) (*ServiceBrowser, error) {
	c := new(ServiceBrowser)

	c.object = conn.Object("org.freedesktop.Avahi", path)
	c.AddChannel = make(chan Service)
	c.RemoveChannel = make(chan Service)
	c.closeCh = make(chan struct{})

	return c, nil
}

func (c *ServiceBrowser) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.ServiceBrowser", method)
}

func (c *ServiceBrowser) free() {
	if c.closeCh != nil {
		close(c.closeCh)
	}
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *ServiceBrowser) getObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *ServiceBrowser) dispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("ItemNew") || signal.Name == c.interfaceForMember("ItemRemove") {
		var service Service
		err := dbus.Store(signal.Body, &service.Interface, &service.Protocol, &service.Name, &service.Type, &service.Domain, &service.Flags)
		if err != nil {
			return err
		}

		if signal.Name == c.interfaceForMember("ItemNew") {
			select {
			case c.AddChannel <- service:
			case <-c.closeCh:
				close(c.AddChannel)
				close(c.RemoveChannel)
				c.closeCh = nil
			}
		} else {
			select {
			case c.RemoveChannel <- service:
			case <-c.closeCh:
				close(c.AddChannel)
				close(c.RemoveChannel)
				c.closeCh = nil
			}
		}
	}

	return nil
}
