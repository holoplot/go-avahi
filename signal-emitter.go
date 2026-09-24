package avahi

import (
	"sync"

	dbus "github.com/godbus/dbus/v5"
)

type signalEmitter interface {
	dispatchSignal(signal *dbus.Signal) error
	getObjectPath() dbus.ObjectPath
	free()
}

// channelCloser closes the channels a signal emitter hands out to its
// consumers without racing a dispatchSignal() that is still sending on them.
type channelCloser struct {
	closeCh chan struct{}
	once    sync.Once
	mutex   sync.Mutex
	closed  bool
}

func newChannelCloser() channelCloser {
	return channelCloser{closeCh: make(chan struct{})}
}

// closeChannels unblocks a pending send, waits for it to return and then
// calls closeFn. Subsequent calls do nothing.
func (c *channelCloser) closeChannels(closeFn func()) {
	c.once.Do(func() {
		close(c.closeCh)

		c.mutex.Lock()
		defer c.mutex.Unlock()

		c.closed = true
		closeFn()
	})
}

// send delivers v on ch, unless the channels are or get closed.
func send[T any](c *channelCloser, ch chan T, v T) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.closed {
		return
	}

	select {
	case ch <- v:
	case <-c.closeCh:
	}
}

func (c *Server) signalEmitterFree(e signalEmitter) {
	o := e.getObjectPath()

	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, ok := c.signalEmitters[o]
	if ok {
		delete(c.signalEmitters, o)
	}

	e.free()
}
