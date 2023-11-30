package avahi

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
)

// A ServiceResolver resolves mDNS services to IP addresses
type ServiceResolver struct {
	object       dbus.BusObject
	FoundChannel chan Service
	closeCh      chan struct{}
}

// ServiceResolverNew returns a new mDNS service resolver
func ServiceResolverNew(conn *dbus.Conn, path dbus.ObjectPath) (*ServiceResolver, error) {
	c := new(ServiceResolver)

	c.object = conn.Object("org.freedesktop.Avahi", path)
	c.FoundChannel = make(chan Service)
	c.closeCh = make(chan struct{})

	return c, nil
}

func (c *ServiceResolver) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.ServiceResolver", method)
}

func (c *ServiceResolver) free() {
	close(c.closeCh)
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *ServiceResolver) getObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *ServiceResolver) dispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("Found") {
		var service Service
		err := dbus.Store(signal.Body, &service.Interface, &service.Protocol,
			&service.Name, &service.Type, &service.Domain, &service.Host,
			&service.Aprotocol, &service.Address, &service.Port,
			&service.Txt, &service.Flags)
		if err != nil {
			return err
		}

		select {
		case c.FoundChannel <- service:
		case <-c.closeCh:
			if !c.isChannelClosed(c.FoundChannel) {
				close(c.FoundChannel)
			}
		}
	}

	return nil
}

// check if a provided channel is closed
func (c *ServiceResolver) isChannelClosed(ch <-chan Service) bool {
	select {
	case <-ch:
		return false
	default:
		return true
	}
}
