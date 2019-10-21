package avahi

import (
	"fmt"

	dbus "github.com/godbus/dbus/v5"
)

const (
	// EntryGroupUncommited - The group has not yet been commited, the user must still call Commit()
	EntryGroupUncommited = 0
	// EntryGroupRegistering - The entries of the group are currently being registered
	EntryGroupRegistering = 1
	// EntryGroupEstablished - The entries have successfully been established
	EntryGroupEstablished = 2
	// EntryGroupCollision - A name collision for one of the entries in the group has been detected, the entries have been withdrawn
	EntryGroupCollision = 3
	// EntryGroupFailure - Some kind of failure happened, the entries have been withdrawn
	EntryGroupFailure = 4
)

// An EntryGroupState describes the current state of an entry group
type EntryGroupState struct {
	State int32
	Error string
}

// An EntryGroup describes a group of records for services
type EntryGroup struct {
	conn               *dbus.Conn
	object             dbus.BusObject
	StateChangeChannel chan EntryGroupState
}

// EntryGroupNew creates a new entry group
func EntryGroupNew(conn *dbus.Conn, path dbus.ObjectPath) (*EntryGroup, error) {
	c := new(EntryGroup)
	c.conn = conn
	c.object = c.conn.Object("org.freedesktop.Avahi", path)
	c.StateChangeChannel = make(chan EntryGroupState, 10)

	return c, nil
}

func (c *EntryGroup) interfaceForMember(method string) string {
	return fmt.Sprintf("%s.%s", "org.freedesktop.Avahi.EntryGroup", method)
}

// Commit an AvahiEntryGroup. The entries in the entry group are now registered on the network.
// Commiting empty entry groups is considered an error.
func (c *EntryGroup) Commit() error {
	return c.object.Call(c.interfaceForMember("Commit"), 0).Err
}

// Reset an AvahiEntryGroup. This takes effect immediately.
func (c *EntryGroup) Reset() error {
	return c.object.Call(c.interfaceForMember("Reset"), 0).Err
}

// GetState gets an AvahiEntryGroup's state
func (c *EntryGroup) GetState() (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetState"), 0).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

// IsEmpty checks if an AvahiEntryGroup is empty
func (c *EntryGroup) IsEmpty() (bool, error) {
	var b bool

	err := c.object.Call(c.interfaceForMember("IsEmpty"), 0).Store(&b)
	if err != nil {
		return false, err
	}

	return b, nil
}

// AddService adds a service. Takes a list of TXT record strings as last arguments.
// Please note that this service is not announced on the network before Commit() is called.
func (c *EntryGroup) AddService(iface, protocol int32, flags uint32, name, serviceType, domain, host string, port uint16, txt [][]byte) error {
	return c.object.Call(c.interfaceForMember("AddService"), 0, iface, protocol, flags, name, serviceType, domain, host, port, txt).Err
}

// AddServiceSubType adds a subtype for a service. The service should already be existent in the entry group.
// You may add as many subtypes for a service as you wish.
func (c *EntryGroup) AddServiceSubType(iface, protocol int32, flags uint32, name, serviceType, domain, subtype string) error {
	return c.object.Call(c.interfaceForMember("AddServiceSubType"), 0, iface, protocol, flags, name, serviceType, domain, subtype).Err
}

// UpdateServiceTxt apdates a TXT record for an existing service.
// The service should already be existent in the entry group.
func (c *EntryGroup) UpdateServiceTxt(iface, protocol int32, flags uint32, name, serviceType, domain string, txt [][]byte) error {
	return c.object.Call(c.interfaceForMember("UpdateServiceTxt"), 0, iface, protocol, flags, name, serviceType, domain, txt).Err
}

// AddAddress add a host/address pair to the entry group
func (c *EntryGroup) AddAddress(iface, protocol int32, flags uint32, name, address string) error {
	return c.object.Call(c.interfaceForMember("AddAddress"), 0, iface, protocol, flags, name, address).Err
}

// AddRecord adds an arbitrary record. I hope you know what you do.
func (c *EntryGroup) AddRecord(iface, protocol int32, flags uint32, name string, class, recordType uint16, ttl uint32, rdata []byte) error {
	return c.object.Call(c.interfaceForMember("AddRecord"), 0, iface, protocol, flags, name, class, recordType, ttl, rdata).Err
}

func (c *EntryGroup) free() {
	c.object.Call(c.interfaceForMember("Free"), 0)
}

func (c *EntryGroup) getObjectPath() dbus.ObjectPath {
	return c.object.Path()
}

func (c *EntryGroup) dispatchSignal(signal *dbus.Signal) error {
	if signal.Name == c.interfaceForMember("StateChanged") {
		var state EntryGroupState
		err := dbus.Store(signal.Body, &state.State, &state.Error)
		if err != nil {
			return err
		}

		c.StateChangeChannel <- state
	}

	return nil
}
