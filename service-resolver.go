package avahi

import (
	"fmt"
	"github.com/godbus/dbus"
)

type ServiceResolver struct {
	object       dbus.BusObject
	FoundChannel chan Service
}

func ServiceResolverNew(conn *dbus.Conn, path dbus.ObjectPath) (*ServiceResolver, error) {
	c := new(ServiceResolver)

	c.object = conn.Object("org.freedesktop.Avahi.ServiceResolver", path)
	c.FoundChannel = make(chan Service, 10)

	return c, nil
}

func (c *ServiceResolver) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.ServiceResolver", method)
}

func (c *ServiceResolver) free() {
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *ServiceResolver) GetObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *ServiceResolver) DispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("Found") {
		var service Service
		err := dbus.Store(signal.Body, &service.Interface, &service.Protocol,
			&service.Name, &service.Type, &service.Domain, &service.Host,
			&service.Aprotocol, &service.Address, &service.Port,
			&service.Txt, &service.Flags)
		if err != nil {
			return err
		}

		c.FoundChannel <- service
	}

	return nil
}
