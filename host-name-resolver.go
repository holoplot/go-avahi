package avahi

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
)

// A HostNameResolver can resolve host names
type HostNameResolver struct {
	object       dbus.BusObject
	FoundChannel chan HostName
}

// HostNameResolverNew returns a new HostNameResolver
func HostNameResolverNew(conn *dbus.Conn, path dbus.ObjectPath) (*HostNameResolver, error) {
	c := new(HostNameResolver)

	c.object = conn.Object("org.freedesktop.Avahi.HostNameResolver", path)
	c.FoundChannel = make(chan HostName, 10)

	return c, nil
}

func (c *HostNameResolver) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.HostNameResolver", method)
}

func (c *HostNameResolver) free() {
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *HostNameResolver) getObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *HostNameResolver) dispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("Found") {
		var hostName HostName
		err := dbus.Store(signal.Body, &hostName.Interface, &hostName.Protocol,
			&hostName.Name, &hostName.Aprotocol, &hostName.Address,
			&hostName.Flags)
		if err != nil {
			return err
		}

		c.FoundChannel <- hostName
	}

	return nil
}
