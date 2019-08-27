package avahi

import (
	"fmt"
	"sync"

	dbus "github.com/godbus/dbus/v5"
)

const (
	SERVER_INVALID     = 0 /* Invalid state (initial) */
	SERVER_REGISTERING = 1 /* Host RRs are being registered */
	SERVER_RUNNING     = 2 /* All host RRs have been established */
	SERVER_COLLISION   = 3 /* There is a collision with a host RR. All host RRs have been withdrawn, the user should set a new host name via avahi_server_set_host_name() */
	SERVER_FAILURE     = 4 /* Some fatal failure happened, the server is unable to proceed */
)

const (
	PROTO_INET   = 0  /* IPv4 */
	PROTO_INET6  = 1  /* IPv6 */
	PROTO_UNSPEC = -1 /* Unspecified/all protocol(s) */
)

const (
	IF_UNSPEC = -1 /* Unspecified/all interface(s) */
)

type SignalEmitter interface {
	dispatchSignal(signal *dbus.Signal) error
	getObjectPath() dbus.ObjectPath
	free()
}

type Server struct {
	conn          *dbus.Conn
	object        dbus.BusObject
	signalChannel chan *dbus.Signal
	quitChannel   chan bool

	mutex          sync.Mutex
	signalEmitters map[dbus.ObjectPath]SignalEmitter
}

type Domain struct {
	Interface int32
	Protocol  int32
	Domain    string
	Flags     uint32
}

type HostName struct {
	Interface int32
	Protocol  int32
	Name      string
	Aprotocol int32
	Address   string
	Flags     uint32
}

type Address struct {
	Interface int32
	Protocol  int32
	Aprotocol int32
	Address   string
	Name      string
	Flags     uint32
}

type ServiceType struct {
	Interface int32
	Protocol  int32
	Type      string
	Domain    string
	Flags     uint32
}

type Service struct {
	Interface int32
	Protocol  int32
	Name      string
	Type      string
	Domain    string
	Host      string
	Aprotocol int32
	Address   string
	Port      int16
	Txt       [][]byte
	Flags     uint32
}

type Record struct {
	Interface int32
	Protocol  int32
	Name      string
	Class     int16
	Type      int16
	Rdata     []byte
	Flags     uint32
}

func ServerNew(conn *dbus.Conn) (*Server, error) {
	c := new(Server)
	c.conn = conn
	c.object = conn.Object("org.freedesktop.Avahi", dbus.ObjectPath("/"))
	c.signalChannel = make(chan *dbus.Signal, 10)
	c.quitChannel = make(chan bool)

	c.conn.Signal(c.signalChannel)
	c.conn.BusObject().Call("org.freedesktop.DBus.AddMatch", 0, "type='signal',interface='org.freedesktop.Avahi.*'")

	c.signalEmitters = make(map[dbus.ObjectPath]SignalEmitter)

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

func (c *Server) Close() {
	<-c.quitChannel

	c.mutex.Lock()
	for path, obj := range c.signalEmitters {
		obj.free()
		delete(c.signalEmitters, path)
	}
	c.mutex.Unlock()

	c.conn.Close()
}

func (c *Server) signalEmitterFree(e SignalEmitter) {
	o := e.getObjectPath()
	e.free()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, ok := c.signalEmitters[o]
	if ok {
		delete(c.signalEmitters, o)
	}
}

func (c *Server) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.Server", method)
}

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

func (c *Server) EntryGroupFree(r *EntryGroup) {
	c.signalEmitterFree(r)
}

func (c *Server) ResolveHostName(interface_ int32, protocol int32, name string, aprotocol int32, flags uint32) (reply HostName, err error) {
	err = c.object.Call(c.interfaceForMember("ResolveHostName"), 0, interface_, protocol, name, aprotocol, flags).
		Store(&reply.Interface, &reply.Protocol, &reply.Name, &reply.Aprotocol, &reply.Address, &reply.Flags)
	return reply, err
}

func (c *Server) ResolveAddress(interface_ int32, protocol int32, address string, flags uint32) (reply Address, err error) {
	err = c.object.Call(c.interfaceForMember("ResolveAddress"), 0, interface_, protocol, address, flags).
		Store(&reply.Interface, &reply.Protocol, &reply.Aprotocol, &reply.Address, &reply.Name, &reply.Flags)
	return reply, err
}

func (c *Server) ResolveService(interface_ int32, protocol int32, name string, type_ string, domain string, aprotocol int32, flags uint32) (reply Service, err error) {
	err = c.object.Call(c.interfaceForMember("ResolveService"), 0, interface_, protocol, name, type_, domain, aprotocol, flags).
		Store(&reply.Interface, &reply.Protocol, &reply.Name, &reply.Type, &reply.Domain,
			&reply.Host, &reply.Aprotocol, &reply.Address, &reply.Port, &reply.Txt, &reply.Flags)
	return reply, err
}

func (c *Server) DomainBrowserNew(interface_ int32, protocol int32, domain string, btype int32, flags uint32) (*DomainBrowser, error) {
	var o dbus.ObjectPath

	err := c.object.Call(c.interfaceForMember("DomainBrowserNew"), 0, interface_, protocol, domain, btype, flags).Store(&o)
	if err != nil {
		return nil, err
	}

	r, err := DomainBrowserNew(c.conn, o)
	if err != nil {
		return nil, err
	}

	return r, nil
}

func (c *Server) DomainBrowserFree(r *DomainBrowser) {
	c.signalEmitterFree(r)
}

func (c *Server) ServiceTypeBrowserNew(interface_ int32, protocol int32, domain string, flags uint32) (*ServiceTypeBrowser, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("ServiceTypeBrowserNew"), 0, interface_, protocol, domain, flags).Store(&o)
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

func (c *Server) ServiceTypeBrowserFree(r *ServiceTypeBrowser) {
	c.signalEmitterFree(r)
}

func (c *Server) ServiceBrowserNew(interface_ int32, protocol int32, type_ string, domain string, flags uint32) (*ServiceBrowser, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("ServiceBrowserNew"), 0, interface_, protocol, type_, domain, flags).Store(&o)
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

func (c *Server) ServiceBrowserFree(r *ServiceBrowser) {
	c.signalEmitterFree(r)
}

func (c *Server) ServiceResolverNew(interface_ int32, protocol int32, name string, type_ string, domain string, aprotocol int32, flags uint32) (*ServiceResolver, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("ServiceResolverNew"), 0, interface_, protocol, name, type_, domain, aprotocol, flags).Store(&o)
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

func (c *Server) ServiceResolverFree(r *ServiceResolver) {
	c.signalEmitterFree(r)
}

func (c *Server) HostNameResolverNew(interface_ int32, protocol int32, name string, aprotocol int32, flags uint32) (*HostNameResolver, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("HostNameResolverNew"), 0, interface_, protocol, name, aprotocol, flags).Store(&o)
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

func (c *Server) AddressResolverNew(interface_ int32, protocol int32, address string, flags uint32) (*AddressResolver, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("AddressResolverNew"), 0, interface_, protocol, address, flags).Store(&o)
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

func (c *Server) AddressResolverFree(r *AddressResolver) {
	c.signalEmitterFree(r)
}

func (c *Server) RecordBrowserNew(interface_ int32, protocol int32, name string, class int16, type_ int16, flags uint32) (*RecordBrowser, error) {
	var o dbus.ObjectPath

	c.mutex.Lock()
	defer c.mutex.Unlock()

	err := c.object.Call(c.interfaceForMember("RecordBrowserNew"), 0, interface_, protocol, name, class, type_, flags).Store(&o)
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

func (c *Server) RecordBrowserFree(r *RecordBrowser) {
	c.signalEmitterFree(r)
}

func (c *Server) GetAPIVersion() (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetAPIVersion"), 0).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (c *Server) GetAlternativeHostName(name string) (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetAlternativeHostName"), 0, name).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *Server) GetAlternativeServiceName(name string) (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetAlternativeServiceName"), 0, name).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *Server) GetDomainName() (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetDomainName"), 0).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *Server) GetHostName() (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetHostName"), 0).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *Server) GetHostNameFqdn() (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetHostNameFqdn"), 0).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *Server) GetLocalServiceCookie() (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetLocalServiceCookie"), 0).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (c *Server) GetNetworkInterfaceIndexByName(name string) (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetNetworkInterfaceIndexByName"), 0, name).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (c *Server) GetNetworkInterfaceNameByIndex(index int32) (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetNetworkInterfaceNameByIndex"), 0, index).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *Server) GetState() (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetState"), 0).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (c *Server) GetVersionString() (string, error) {
	var s string

	err := c.object.Call(c.interfaceForMember("GetVersionString"), 0).Store(&s)
	if err != nil {
		return "", err
	}

	return s, nil
}

func (c *Server) IsNSSSupportAvailable() (bool, error) {
	var b bool

	err := c.object.Call(c.interfaceForMember("IsNSSSupportAvailable"), 0).Store(&b)
	if err != nil {
		return false, err
	}

	return b, nil
}

func (c *Server) SetServerName(name string) error {
	return c.object.Call(c.interfaceForMember("SetServerName"), 0, name).Err
}
