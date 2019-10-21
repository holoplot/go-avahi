package avahi

import (
	"fmt"
	"sync"

	dbus "github.com/godbus/dbus/v5"
)

const (
	// ServerInvalid - Invalid state (initial)
	ServerInvalid = 0
	// ServerRegistering - Host RRs are being registered
	ServerRegistering = 1
	// ServerRunning - All host RRs have been established
	ServerRunning = 2
	// ServerCollision - There is a collision with a host RR. All host RRs have been withdrawn, the user should set a new host name via SetHostname()
	ServerCollision = 3
	// ServerFailure - Some fatal failure happened, the server is unable to proceed
	ServerFailure = 4
)

// A Server is the cental object of an Avahi connection
type Server struct {
	conn          *dbus.Conn
	object        dbus.BusObject
	signalChannel chan *dbus.Signal
	quitChannel   chan struct{}

	mutex          sync.Mutex
	signalEmitters map[dbus.ObjectPath]signalEmitter
}

// ServerNew returns a new Server object
func ServerNew(conn *dbus.Conn) (*Server, error) {
	c := new(Server)
	c.conn = conn
	c.object = conn.Object("org.freedesktop.Avahi", dbus.ObjectPath("/"))
	c.signalChannel = make(chan *dbus.Signal, 10)
	c.quitChannel = make(chan struct{})

	c.conn.Signal(c.signalChannel)
	c.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='org.freedesktop.Avahi.*'")

	c.signalEmitters = make(map[dbus.ObjectPath]signalEmitter)

	go func() {
		for {
			select {
			case signal, ok := <-c.signalChannel:
				if !ok {
					continue
				}

				c.mutex.Lock()
				for path, obj := range c.signalEmitters {
					if path == signal.Path {
						obj.dispatchSignal(signal)
					}
				}
				c.mutex.Unlock()

			case <-c.quitChannel:
				return
			}
		}
	}()

	return c, nil
}

// Close closes the connection to a server
func (c *Server) Close() {
	c.quitChannel <- struct{}{}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	for path, obj := range c.signalEmitters {
		obj.free()
		delete(c.signalEmitters, path)
	}

	c.conn.Close()
}

func (c *Server) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.Server", method)
}

// EntryGroupNew returns a new and empty EntryGroup
func (c *Server) EntryGroupNew() (*EntryGroup, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("EntryGroupNew"), 0).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := EntryGroupNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	c.signalEmitters[o] = r

	return r, nil
}

// EntryGroupFree frees an entry group and releases its resources on the service
func (c *Server) EntryGroupFree(r *EntryGroup) {
	c.signalEmitterFree(r)
}

// ResolveHostName ...
func (c *Server) ResolveHostName(iface, protocol int32, name string, aprotocol int32, flags uint32) (reply HostName, err error) {
	err = c.object.Call(c.interfaceForMember("ResolveHostName"), 0, iface, protocol, name, aprotocol, flags).
		Store(&reply.Interface, &reply.Protocol, &reply.Name, &reply.Aprotocol, &reply.Address, &reply.Flags)
	return reply, err
}

// ResolveAddress ...
func (c *Server) ResolveAddress(iface, protocol int32, address string, flags uint32) (reply Address, err error) {
	err = c.object.Call(c.interfaceForMember("ResolveAddress"), 0, iface, protocol, address, flags).
		Store(&reply.Interface, &reply.Protocol, &reply.Aprotocol, &reply.Address, &reply.Name, &reply.Flags)
	return reply, err
}

// ResolveService ...
func (c *Server) ResolveService(iface, protocol int32, name, serviceType, domain string, aprotocol int32, flags uint32) (reply Service, err error) {
	err = c.object.Call(c.interfaceForMember("ResolveService"), 0, iface, protocol, name, serviceType, domain, aprotocol, flags).
		Store(&reply.Interface, &reply.Protocol, &reply.Name, &reply.Type, &reply.Domain,
			&reply.Host, &reply.Aprotocol, &reply.Address, &reply.Port, &reply.Txt, &reply.Flags)
	return reply, err
}

// DomainBrowserNew ...
func (c *Server) DomainBrowserNew(iface, protocol int32, domain string, btype, flags uint32) (*DomainBrowser, error) {
	var o dbus.ObjectPath

	err := c.object.Call(c.interfaceForMember("DomainBrowserNew"), 0, iface, protocol, domain, btype, flags).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := DomainBrowserNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	return r, nil
}

// DomainBrowserFree ...
func (c *Server) DomainBrowserFree(r *DomainBrowser) {
	c.signalEmitterFree(r)
}

// ServiceTypeBrowserNew ...
func (c *Server) ServiceTypeBrowserNew(iface, protocol int32, domain string, flags uint32) (*ServiceTypeBrowser, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("ServiceTypeBrowserNew"), 0, iface, protocol, domain, flags).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := ServiceTypeBrowserNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	c.signalEmitters[o] = r

	return r, nil
}

// ServiceTypeBrowserFree ...
func (c *Server) ServiceTypeBrowserFree(r *ServiceTypeBrowser) {
	c.signalEmitterFree(r)
}

// ServiceBrowserNew ...
func (c *Server) ServiceBrowserNew(iface, protocol int32, serviceType string, domain string, flags uint32) (*ServiceBrowser, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("ServiceBrowserNew"), 0, iface, protocol, serviceType, domain, flags).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := ServiceBrowserNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	c.signalEmitters[o] = r

	return r, nil
}

// ServiceBrowserFree ...
func (c *Server) ServiceBrowserFree(r *ServiceBrowser) {
	c.signalEmitterFree(r)
}

// ServiceResolverNew ...
func (c *Server) ServiceResolverNew(iface, protocol int32, name, serviceType, domain string, aprotocol int32, flags uint32) (*ServiceResolver, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("ServiceResolverNew"), 0, iface, protocol, name, serviceType, domain, aprotocol, flags).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := ServiceResolverNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	c.signalEmitters[o] = r

	return r, nil
}

// ServiceResolverFree ...
func (c *Server) ServiceResolverFree(r *ServiceResolver) {
	c.signalEmitterFree(r)
}

// HostNameResolverNew ...
func (c *Server) HostNameResolverNew(iface, protocol int32, name string, aprotocol int32, flags uint32) (*HostNameResolver, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("HostNameResolverNew"), 0, iface, protocol, name, aprotocol, flags).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := HostNameResolverNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	c.signalEmitters[o] = r

	return r, nil
}

// AddressResolverNew ...
func (c *Server) AddressResolverNew(iface, protocol int32, address string, flags uint32) (*AddressResolver, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("AddressResolverNew"), 0, iface, protocol, address, flags).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := AddressResolverNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	c.signalEmitters[o] = r

	return r, nil
}

// AddressResolverFree ...
func (c *Server) AddressResolverFree(r *AddressResolver) {
	c.signalEmitterFree(r)
}

// RecordBrowserNew ...
func (c *Server) RecordBrowserNew(iface, protocol int32, name string, class int16, recordType int16, flags uint32) (*RecordBrowser, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("RecordBrowserNew"), 0, iface, protocol, name, class, recordType, flags).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := RecordBrowserNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	c.signalEmitters[o] = r

	return r, nil
}

// RecordBrowserFree ...
func (c *Server) RecordBrowserFree(r *RecordBrowser) {
	c.signalEmitterFree(r)
}

// GetAPIVersion ...
func (c *Server) GetAPIVersion() (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetAPIVersion"), 0).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetAlternativeHostName ...
func (c *Server) GetAlternativeHostName(name string) (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetAlternativeHostName"), 0, name).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

// GetAlternativeServiceName ...
func (c *Server) GetAlternativeServiceName(name string) (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetAlternativeServiceName"), 0, name).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

// GetDomainName ...
func (c *Server) GetDomainName() (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetDomainName"), 0).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

// GetHostName ...
func (c *Server) GetHostName() (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetHostName"), 0).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

// GetHostNameFqdn ...
func (c *Server) GetHostNameFqdn() (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetHostNameFqdn"), 0).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

// GetLocalServiceCookie ...
func (c *Server) GetLocalServiceCookie() (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetLocalServiceCookie"), 0).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetNetworkInterfaceIndexByName -...
func (c *Server) GetNetworkInterfaceIndexByName(name string) (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetNetworkInterfaceIndexByName"), 0, name).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetNetworkInterfaceNameByIndex ...
func (c *Server) GetNetworkInterfaceNameByIndex(index int32) (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetNetworkInterfaceNameByIndex"), 0, index).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

// GetState ...
func (c *Server) GetState() (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetState"), 0).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// GetVersionString ...
func (c *Server) GetVersionString() (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetVersionString"), 0).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

// IsNSSSupportAvailable ...
func (c *Server) IsNSSSupportAvailable() (bool, error) {
	var b bool

	err := c.object.Call(c.interfaceForMember("IsNSSSupportAvailable"), 0).Store(&b)
	if err != nil {
		return false, err
	}

	return b, nil
}

// SetServerName ...
func (c *Server) SetServerName(name string) error {
	return c.object.Call(c.interfaceForMember("SetServerName"), 0, name).Err
}
