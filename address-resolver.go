package avahi

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
)

// An AddressResolver resolves Address to IP addresses
type AddressResolver struct {
	object       dbus.BusObject
	FoundChannel chan Address
	closeCh      chan struct{}
}

// AddressResolverNew creates a new AddressResolver
func AddressResolverNew(conn *dbus.Conn, path dbus.ObjectPath) (*AddressResolver, error) {
	c := new(AddressResolver)

	c.object = conn.Object("org.freedesktop.Avahi", path)
	c.FoundChannel = make(chan Address)
	c.closeCh = make(chan struct{})

	return c, nil
}

func (c *AddressResolver) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.AddressResolver", method)
}

func (c *AddressResolver) free() {
	if c.closeCh != nil {
		close(c.closeCh)
	}
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *AddressResolver) getObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *AddressResolver) dispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("Found") {
		var address Address
		err := dbus.Store(signal.Body, &address.Interface, &address.Protocol,
			&address.Aprotocol, &address.Address, &address.Name,
			&address.Flags)
		if err != nil {
			return err
		}

		select {
		case c.FoundChannel <- address:
		case <-c.closeCh:
			close(c.FoundChannel)
			c.closeCh = nil
		}
	}

	return nil
}
