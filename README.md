# Golang bindings for Avahi

Avahi is an implementation of the mDNS protocol. Refer to the [Wikipedia article](https://en.wikipedia.org/wiki/Avahi_(software)),
the [website](https://www.avahi.org/) and the [GitHub project](https://github.com/lathiat/avahi) for further information.

This Go package provides bindings for DBus interfaces exposed by the Avahi daemon.

# Install

Install the package like this:

```
go get github.com/holoplot/go-avahi
```

And then use it in your source code.

```
import "github.com/holoplot/go-avahi"
```

# Examples

Below are some examples to illustrate the usage of this package.
Note that you will need to have a working Avahi installation.

## Browsing and resolving

```go
package main

import (
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("Cannot get system bus: %v", err)
	}

	server, err := avahi.ServerNew(conn)
	if err != nil {
		log.Fatalf("Avahi new failed: %v", err)
	}

	host, err := server.GetHostName()
	if err != nil {
		log.Fatalf("GetHostName() failed: %v", err)
	}
	log.Println("GetHostName()", host)

	fqdn, err := server.GetHostNameFqdn()
	if err != nil {
		log.Fatalf("GetHostNameFqdn() failed: %v", err)
	}
	log.Println("GetHostNameFqdn()", fqdn)

	s, err := server.GetAlternativeHostName(host)
	if err != nil {
		log.Fatalf("GetAlternativeHostName() failed: %v", err)
	}
	log.Println("GetAlternativeHostName()", s)

	i, err := server.GetAPIVersion()
	if err != nil {
		log.Fatalf("GetAPIVersion() failed: %v", err)
	}
	log.Println("GetAPIVersion()", i)

	hn, err := server.ResolveHostName(avahi.InterfaceUnspec, avahi.ProtoUnspec, fqdn, avahi.ProtoUnspec, 0)
	if err != nil {
		log.Fatalf("ResolveHostName() failed: %v", err)
	}
	log.Println("ResolveHostName:", hn)

	db, err := server.DomainBrowserNew(avahi.InterfaceUnspec, avahi.ProtoUnspec, "", avahi.DomainBrowserTypeBrowseDefault, 0)
	if err != nil {
		log.Fatalf("DomainBrowserNew() failed: %v", err)
	}

	stb, err := server.ServiceTypeBrowserNew(avahi.InterfaceUnspec, avahi.ProtoUnspec, "local", 0)
	if err != nil {
		log.Fatalf("ServiceTypeBrowserNew() failed: %v", err)
	}

	sb, err := server.ServiceBrowserNew(avahi.InterfaceUnspec, avahi.ProtoUnspec, "_my-nifty-service._tcp", "local", 0)
	if err != nil {
		log.Fatalf("ServiceBrowserNew() failed: %v", err)
	}

	sr, err := server.ServiceResolverNew(avahi.InterfaceUnspec, avahi.ProtoUnspec, "", "_my-nifty-service._tcp", "local", avahi.ProtoUnspec, 0)
	if err != nil {
		log.Fatalf("ServiceResolverNew() failed: %v", err)
	}

	var domain avahi.Domain
	var service avahi.Service
	var serviceType avahi.ServiceType

	for {
		select {
		case domain = <-db.AddChannel:
			log.Println("DomainBrowser ADD: ", domain)
		case domain = <-db.RemoveChannel:
			log.Println("DomainBrowser REMOVE: ", domain)
		case serviceType = <-stb.AddChannel:
			log.Println("ServiceTypeBrowser ADD: ", serviceType)
		case serviceType = <-stb.RemoveChannel:
			log.Println("ServiceTypeBrowser REMOVE: ", serviceType)
		case service = <-sb.AddChannel:
			log.Println("ServiceBrowser ADD: ", service)

			service, err := server.ResolveService(service.Interface, service.Protocol, service.Name,
				service.Type, service.Domain, avahi.ProtoUnspec, 0)
			if err == nil {
				log.Println(" RESOLVED >>", service.Address)
			}
		case service = <-sb.RemoveChannel:
			log.Println("ServiceBrowser REMOVE: ", service)
		case service = <-sr.FoundChannel:
			log.Println("ServiceResolver FOUND: ", service)
		}
	}
}
```

## Publishing

```go
package main

import (
	"log"

	"github.com/godbus/dbus/v5"
	"github.com/holoplot/go-avahi"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatalf("Cannot get system bus: %v", err)
	}

	a, err := avahi.ServerNew(conn)
	if err != nil {
		log.Fatalf("Avahi new failed: %v", err)
	}

	eg, err := a.EntryGroupNew()
	if err != nil {
		log.Fatalf("EntryGroupNew() failed: %v", err)
	}

	hostname, err := a.GetHostName()
	if err != nil {
		log.Fatalf("GetHostName() failed: %v", err)
	}

	fqdn, err := a.GetHostNameFqdn()
	if err != nil {
		log.Fatalf("GetHostNameFqdn() failed: %v", err)
	}

	err = eg.AddService(avahi.InterfaceUnspec, avahi.ProtoUnspec, 0, hostname, "_my-nifty-service._tcp", "local", fqdn, 1234, nil)
	if err != nil {
		log.Fatalf("AddService() failed: %v", err)
	}

	err = eg.Commit()
	if err != nil {
		log.Fatalf("Commit() failed: %v", err)
	}

	log.Println("Entry published. Hit ^C to exit.")

	for {
		select {}
	}
}
```

# MIT License

See file `LICENSE` for details.
