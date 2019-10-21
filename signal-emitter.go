package avahi

import dbus "github.com/godbus/dbus/v5"

type signalEmitter interface {
	dispatchSignal(signal *dbus.Signal) error
	getObjectPath() dbus.ObjectPath
	free()
}

func (c *Server) signalEmitterFree(e signalEmitter) {
	o := e.getObjectPath()
	e.free()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, ok := c.signalEmitters[o]
	if ok {
		delete(c.signalEmitters, o)
	}
}
