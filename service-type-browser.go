package avahi

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
)

// A ServiceTypeBrowser is used to browser the mDNS network for services of a specific type
type ServiceTypeBrowser struct {
	object        dbus.BusObject
	AddChannel    chan ServiceType
	RemoveChannel chan ServiceType
}

// ServiceTypeBrowserNew creates a new browser for mDNS service types
func ServiceTypeBrowserNew(conn *dbus.Conn, path dbus.ObjectPath) (*ServiceTypeBrowser, error) {
	c := new(ServiceTypeBrowser)

	c.object = conn.Object("org.freedesktop.Avahi.ServiceTypeBrowser", path)
	c.AddChannel = make(chan ServiceType, 10)
	c.RemoveChannel = make(chan ServiceType, 10)

	return c, nil
}

func (c *ServiceTypeBrowser) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.ServiceTypeBrowser", method)
}

func (c *ServiceTypeBrowser) free() {
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *ServiceTypeBrowser) getObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *ServiceTypeBrowser) dispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("ItemNew") || signal.Name == c.interfaceForMember("ItemRemove") {
		var serviceType ServiceType
		err := dbus.Store(signal.Body, &serviceType.Interface, &serviceType.Protocol, &serviceType.Type, &serviceType.Domain, &serviceType.Flags)
		if err != nil {
			return err
		}

		if signal.Name == c.interfaceForMember("ItemNew") {
			c.AddChannel <- serviceType
		} else {
			c.RemoveChannel <- serviceType
		}
	}

	return nil
}
