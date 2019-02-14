package avahi

import (
	"fmt"

	"github.com/godbus/dbus"
)

const (
	ENTRY_GROUP_UNCOMMITED  = 0 /* The group has not yet been commited, the user must still call avahi_entry_group_commit() */
	ENTRY_GROUP_REGISTERING = 1 /* The entries of the group are currently being registered */
	ENTRY_GROUP_ESTABLISHED = 2 /* The entries have successfully been established */
	ENTRY_GROUP_COLLISION   = 3 /* A name collision for one of the entries in the group has been detected, the entries have been withdrawn */
	ENTRY_GROUP_FAILURE     = 4 /* Some kind of failure happened, the entries have been withdrawn */
)

type EntryGroupState struct {
	State int32
	Error string
}

type EntryGroup struct {
	conn               *dbus.Conn
	object             dbus.BusObject
	StateChangeChannel chan EntryGroupState
}

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

func (c *EntryGroup) Commit() error {
	return c.object.Call(c.interfaceForMember("Commit"), 0).Err
}

func (c *EntryGroup) Reset() error {
	return c.object.Call(c.interfaceForMember("Reset"), 0).Err
}

func (c *EntryGroup) GetState() (int32, error) {
	var i int32

	err := c.object.Call(c.interfaceForMember("GetState"), 0).Store(&i)
	if err != nil {
		return 0, err
	}

	return i, nil
}

func (c *EntryGroup) IsEmpty() (bool, error) {
	var b bool

	err := c.object.Call(c.interfaceForMember("IsEmpty"), 0).Store(&b)
	if err != nil {
		return false, err
	}

	return b, nil
}

func (c *EntryGroup) AddService(interface_ int32, protocol int32, flags uint32, name string, type_ string, domain string, host string, port uint16, txt [][]byte) error {
	return c.object.Call(c.interfaceForMember("AddService"), 0, interface_, protocol, flags, name, type_, domain, host, port, txt).Err
}

func (c *EntryGroup) AddServiceSubType(interface_ int32, protocol int32, flags uint32, name string, type_ string, domain string, subtype string) error {
	return c.object.Call(c.interfaceForMember("AddServiceSubType"), 0, interface_, protocol, flags, name, type_, domain, subtype).Err
}

func (c *EntryGroup) UpdateServiceTxt(interface_ int32, protocol int32, flags uint32, name string, type_ string, domain string, txt [][]byte) error {
	return c.object.Call(c.interfaceForMember("UpdateServiceTxt"), 0, interface_, protocol, flags, name, type_, domain, txt).Err
}

func (c *EntryGroup) AddAddress(interface_ int32, protocol int32, flags uint32, name string, address string) error {
	return c.object.Call(c.interfaceForMember("AddAddress"), 0, interface_, protocol, flags, name, address).Err
}

func (c *EntryGroup) AddRecord(interface_ int32, protocol int32, flags uint32, name string, class uint16, type_ uint16, ttl uint32, rdata []byte) error {
	return c.object.Call(c.interfaceForMember("AddRecord"), 0, interface_, protocol, flags, name, class, type_, ttl, rdata).Err
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
