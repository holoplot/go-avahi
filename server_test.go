package avahi

import (
	"github.com/godbus/dbus"
	"testing"
)

// TestNew ensures that New() works without errors.
func TestNew(t *testing.T) {
	conn, err := dbus.SystemBus()
	if err != nil {
		t.Fatal(err)
	}

	_, err = ServerNew(conn)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBasic(t *testing.T) {
	conn, err := dbus.SystemBus()
	if err != nil {
		t.Fatal(err)
	}

	a, err := ServerNew(conn)
	if err != nil {
		t.Fatal("Avahi new failed")
	}

	s, err := a.GetHostName()
	if err != nil {
		t.Fatal("GetHostName() failed")
	}
	t.Log("GetHostName()", s)

	s, err = a.GetAlternativeHostName(s)
	if err != nil {
		t.Fatal("GetAlternativeHostName() failed")
	}
	t.Log("GetAlternativeHostName()", s)

	////

	i, err := a.GetAPIVersion()
	if err != nil {
		t.Fatal("GetAPIVersion() failed")
	}
	t.Log("GetAPIVersion()", i)

	s, err = a.GetNetworkInterfaceNameByIndex(1)
	if err != nil {
		t.Fatal("GetNetworkInterfaceNameByIndex() failed")
	}
	t.Log("GetNetworkInterfaceNameByIndex()", s)

	i, err = a.GetNetworkInterfaceIndexByName(s)
	if err != nil {
		t.Fatal("GetNetworkInterfaceIndexByName() failed")
	}
	if i != 1 {
		t.Fatal("GetNetworkInterfaceIndexByName() returned wrong index")
	}
	t.Log("GetNetworkInterfaceIndexByName()", i)

	///

	egc, err := a.EntryGroupNew()
	if err != nil {
		t.Fatal("EntryGroupNew() failed")
	}

	b, err := egc.IsEmpty()
	if err != nil {
		t.Fatal("egc.IsEmpty() failed")
	}
	if b != true {
		t.Fatal("Entry group must initially be empty")
	}

	egc.Free()

}
