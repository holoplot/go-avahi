package avahi

import (
	"reflect"
	"testing"
	"time"

	dbus "github.com/godbus/dbus/v5"
)

// fakeBusObject stands in for the Avahi D-Bus object so that free() can
// be exercised without a running daemon.
type fakeBusObject struct {
	dbus.BusObject
}

func (fakeBusObject) Call(string, dbus.Flags, ...any) *dbus.Call {
	return &dbus.Call{}
}

func (fakeBusObject) Path() dbus.ObjectPath {
	return "/test"
}

type emitterTestCase struct {
	name string
	// new returns a fresh emitter with its D-Bus object replaced by a fake,
	// and the channels its consumers read from.
	new func() (signalEmitter, []any)
	// signals returns one valid signal per member the emitter forwards.
	signals func(e signalEmitter) []*dbus.Signal
	// buffered is the capacity of the emitter's channels.
	buffered int
}

func signal(e signalEmitter, iface, member string, body ...any) *dbus.Signal {
	return &dbus.Signal{
		Path: e.getObjectPath(),
		Name: "org.freedesktop.Avahi." + iface + "." + member,
		Body: body,
	}
}

var emitterTestCases = []emitterTestCase{
	{
		name: "AddressResolver",
		new: func() (signalEmitter, []any) {
			c, _ := AddressResolverNew(nil, "/test")
			c.object = fakeBusObject{}
			return c, []any{c.FoundChannel}
		},
		signals: func(e signalEmitter) []*dbus.Signal {
			return []*dbus.Signal{
				signal(e, "AddressResolver", "Found", int32(0), int32(0), int32(0), "a", "n", uint32(0)),
			}
		},
	},
	{
		name: "DomainBrowser",
		new: func() (signalEmitter, []any) {
			c, _ := DomainBrowserNew(nil, "/test")
			c.object = fakeBusObject{}
			return c, []any{c.AddChannel, c.RemoveChannel}
		},
		signals: func(e signalEmitter) []*dbus.Signal {
			return []*dbus.Signal{
				signal(e, "DomainBrowser", "ItemNew", int32(0), int32(0), "d", uint32(0)),
				signal(e, "DomainBrowser", "ItemRemove", int32(0), int32(0), "d", uint32(0)),
			}
		},
	},
	{
		name: "EntryGroup",
		new: func() (signalEmitter, []any) {
			c, _ := EntryGroupNew(nil, "/test")
			c.object = fakeBusObject{}
			return c, []any{c.StateChangeChannel}
		},
		signals: func(e signalEmitter) []*dbus.Signal {
			return []*dbus.Signal{
				signal(e, "EntryGroup", "StateChanged", int32(0), ""),
			}
		},
		buffered: 10,
	},
	{
		name: "HostNameResolver",
		new: func() (signalEmitter, []any) {
			c, _ := HostNameResolverNew(nil, "/test")
			c.object = fakeBusObject{}
			return c, []any{c.FoundChannel}
		},
		signals: func(e signalEmitter) []*dbus.Signal {
			return []*dbus.Signal{
				signal(e, "HostNameResolver", "Found", int32(0), int32(0), "n", int32(0), "a", uint32(0)),
			}
		},
	},
	{
		name: "RecordBrowser",
		new: func() (signalEmitter, []any) {
			c, _ := RecordBrowserNew(nil, "/test")
			c.object = fakeBusObject{}
			return c, []any{c.AddChannel, c.RemoveChannel}
		},
		signals: func(e signalEmitter) []*dbus.Signal {
			return []*dbus.Signal{
				signal(e, "RecordBrowser", "ItemNew", int32(0), int32(0), "n", uint16(0), uint16(0), []byte{}, uint32(0)),
				signal(e, "RecordBrowser", "ItemRemove", int32(0), int32(0), "n", uint16(0), uint16(0), []byte{}, uint32(0)),
			}
		},
	},
	{
		name: "ServiceBrowser",
		new: func() (signalEmitter, []any) {
			c, _ := ServiceBrowserNew(nil, "/test")
			c.object = fakeBusObject{}
			return c, []any{c.AddChannel, c.RemoveChannel}
		},
		signals: func(e signalEmitter) []*dbus.Signal {
			return []*dbus.Signal{
				signal(e, "ServiceBrowser", "ItemNew", int32(0), int32(0), "n", "t", "d", uint32(0)),
				signal(e, "ServiceBrowser", "ItemRemove", int32(0), int32(0), "n", "t", "d", uint32(0)),
			}
		},
	},
	{
		name: "ServiceResolver",
		new: func() (signalEmitter, []any) {
			c, _ := ServiceResolverNew(nil, "/test")
			c.object = fakeBusObject{}
			return c, []any{c.FoundChannel}
		},
		signals: func(e signalEmitter) []*dbus.Signal {
			return []*dbus.Signal{
				signal(e, "ServiceResolver", "Found", int32(0), int32(0), "n", "t", "d", "h",
					int32(0), "a", uint16(0), [][]byte{}, uint32(0)),
			}
		},
	},
	{
		name: "ServiceTypeBrowser",
		new: func() (signalEmitter, []any) {
			c, _ := ServiceTypeBrowserNew(nil, "/test")
			c.object = fakeBusObject{}
			return c, []any{c.AddChannel, c.RemoveChannel}
		},
		signals: func(e signalEmitter) []*dbus.Signal {
			return []*dbus.Signal{
				signal(e, "ServiceTypeBrowser", "ItemNew", int32(0), int32(0), "t", "d", uint32(0)),
				signal(e, "ServiceTypeBrowser", "ItemRemove", int32(0), int32(0), "t", "d", uint32(0)),
			}
		},
	},
}

// within fails the test if fn does not return in time.
func within(t *testing.T, what string, fn func()) {
	t.Helper()

	done := make(chan struct{})
	go func() {
		fn()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("%s did not return", what)
	}
}

// assertClosed drains ch and fails the test unless it is closed.
func assertClosed(t *testing.T, ch any) {
	t.Helper()

	v := reflect.ValueOf(ch)
	within(t, "draining channel", func() {
		for {
			if _, ok := v.Recv(); !ok {
				return
			}
		}
	})
}

func dispatch(t *testing.T, e signalEmitter, s *dbus.Signal) {
	t.Helper()

	if err := e.dispatchSignal(s); err != nil {
		t.Errorf("dispatchSignal(%s): %v", s.Name, err)
	}
}

func TestSignalEmitterFreeClosesChannels(t *testing.T) {
	for _, tc := range emitterTestCases {
		t.Run(tc.name, func(t *testing.T) {
			e, channels := tc.new()

			e.free()

			for _, ch := range channels {
				assertClosed(t, ch)
			}
		})
	}
}

func TestSignalEmitterFreeTwice(t *testing.T) {
	for _, tc := range emitterTestCases {
		t.Run(tc.name, func(t *testing.T) {
			e, _ := tc.new()

			e.free()
			e.free()
		})
	}
}

func TestSignalEmitterDispatchAfterFree(t *testing.T) {
	for _, tc := range emitterTestCases {
		t.Run(tc.name, func(t *testing.T) {
			e, _ := tc.new()

			e.free()

			within(t, "dispatchSignal after free", func() {
				for i := 0; i <= tc.buffered; i++ {
					for _, s := range tc.signals(e) {
						dispatch(t, e, s)
					}
				}
			})
		})
	}
}

func TestSignalEmitterFreeWithPendingDispatch(t *testing.T) {
	for _, tc := range emitterTestCases {
		t.Run(tc.name, func(t *testing.T) {
			probe, _ := tc.new()

			for k := range tc.signals(probe) {
				for i := 0; i < 100; i++ {
					e, channels := tc.new()
					s := tc.signals(e)[k]

					// Fill any buffer so that the next dispatch blocks.
					for j := 0; j < tc.buffered; j++ {
						dispatch(t, e, s)
					}

					dispatched := make(chan struct{})
					go func() {
						dispatch(t, e, s)
						close(dispatched)
					}()

					if i%2 == 0 {
						time.Sleep(time.Millisecond)
					}

					within(t, "free", e.free)
					within(t, "pending dispatchSignal", func() { <-dispatched })

					for _, ch := range channels {
						assertClosed(t, ch)
					}
				}
			}
		})
	}
}

func TestServerFreeWithPendingDispatch(t *testing.T) {
	for _, tc := range emitterTestCases {
		t.Run(tc.name, func(t *testing.T) {
			e, channels := tc.new()
			s := &Server{
				signalEmitters: map[dbus.ObjectPath]signalEmitter{
					e.getObjectPath(): e,
				},
			}
			sig := tc.signals(e)[0]

			// Fill any buffer so that the next dispatch blocks.
			for j := 0; j < tc.buffered; j++ {
				s.dispatchSignal(sig)
			}

			dispatched := make(chan struct{})
			go func() {
				s.dispatchSignal(sig)
				close(dispatched)
			}()

			time.Sleep(time.Millisecond)

			within(t, "signalEmitterFree", func() { s.signalEmitterFree(e) })
			within(t, "pending dispatchSignal", func() { <-dispatched })

			for _, ch := range channels {
				assertClosed(t, ch)
			}
		})
	}
}

func TestServerCloseWithPendingDispatch(t *testing.T) {
	for _, tc := range emitterTestCases {
		t.Run(tc.name, func(t *testing.T) {
			e, channels := tc.new()
			s := &Server{
				object:        fakeBusObject{},
				signalChannel: make(chan *dbus.Signal, 10),
				quitChannel:   make(chan struct{}),
				signalEmitters: map[dbus.ObjectPath]signalEmitter{
					e.getObjectPath(): e,
				},
			}
			go s.dispatchLoop()

			// Fill any buffer, then queue one signal that blocks the loop.
			for j := 0; j < tc.buffered; j++ {
				s.dispatchSignal(tc.signals(e)[0])
			}
			s.signalChannel <- tc.signals(e)[0]

			time.Sleep(time.Millisecond)

			within(t, "Close", s.Close)

			for _, ch := range channels {
				assertClosed(t, ch)
			}
		})
	}
}
