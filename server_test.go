package avahi

import (
	"testing"
	"time"

	"github.com/godbus/dbus/v5"
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

func TestNewClose(t *testing.T) {
	conn, err := dbus.SystemBus()
	if err != nil {
		t.Fatal(err)
	}

	a, err := ServerNew(conn)
	if err != nil {
		t.Fatal(err)
	}

	doneChannel := make(chan struct{})
	go func() {
		a.Close()
		doneChannel <- struct{}{}
	}()

	select {
	case <-time.After(2 * time.Second):
		t.Fatal("Close() is deadlocked")
	case <-doneChannel:
	}
}

func TestBasic(t *testing.T) {
	conn, err := dbus.SystemBus()
	if err != nil {
		t.Fatal(err)
	}

	a, err := ServerNew(conn)
	if err != nil {
		t.Fatalf("Avahi new failed: %v", err)
	}

	s, err := a.GetHostName()
	if err != nil {
		t.Fatalf("GetHostName() failed: %v", err)
	}
	t.Log("GetHostName()", s)

	s, err = a.GetAlternativeHostName(s)
	if err != nil {
		t.Fatalf("GetAlternativeHostName() failed: %v", err)
	}
	t.Log("GetAlternativeHostName()", s)

	////

	i, err := a.GetAPIVersion()
	if err != nil {
		t.Fatalf("GetAPIVersion() failed: %v", err)
	}
	t.Log("GetAPIVersion()", i)

	s, err = a.GetNetworkInterfaceNameByIndex(1)
	if err != nil {
		t.Fatalf("GetNetworkInterfaceNameByIndex() failed: %v", err)
	}
	t.Log("GetNetworkInterfaceNameByIndex()", s)

	i, err = a.GetNetworkInterfaceIndexByName(s)
	if err != nil {
		t.Fatalf("GetNetworkInterfaceIndexByName() failed: %v", err)
	}
	if i != 1 {
		t.Fatal("GetNetworkInterfaceIndexByName() returned wrong index")
	}
	t.Log("GetNetworkInterfaceIndexByName()", i)

	///

	egc, err := a.EntryGroupNew()
	if err != nil {
		t.Fatalf("EntryGroupNew() failed: %v", err)
	}

	b, err := egc.IsEmpty()
	if err != nil {
		t.Fatalf("egc.IsEmpty() failed: %v", err)
	}
	if b != true {
		t.Fatal("Entry group must initially be empty")
	}
}
