# Golang bindings for Avahi

Avahi is an implementation of the mDNS protocol. Refer to the [Wikipedia article](https://en.wikipedia.org/wiki/Avahi_(software)),
the [website](https://www.avahi.org/) and the [GitHub project](https://github.com/lathiat/avahi) for further information.

This Go package provides bindings for DBus interfaces exposed by the Avahi daemon.

# Install

Install the package like this:

```
go get https://github.com/holoplot/go-avahi
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
	"github.com/godbus/dbus"
	"guthub.com/holoplot/go-avahi"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatal("Cannot get system bus")
	}

	server, err := avahi.ServerNew(conn)
	if err != nil {
		log.Fatal("Avahi new failed")
	}

	host, err := server.GetHostName()
	if err != nil {
		log.Fatal("GetHostName() failed")
	}
	log.Println("GetHostName()", host)

	fqdn, err := server.GetHostNameFqdn()
	if err != nil {
		log.Fatal("GetHostNameFqdn() failed")
	}
	log.Println("GetHostNameFqdn()", fqdn)

	s, err := server.GetAlternativeHostName(host)
	if err != nil {
		log.Fatal("GetAlternativeHostName() failed")
	}
	log.Println("GetAlternativeHostName()", s)


	i, err := server.GetAPIVersion()
	if err != nil {
		log.Fatal("GetAPIVersion() failed")
	}
	log.Println("GetAPIVersion()", i)

	hn, err := server.ResolveHostName(avahi.IF_UNSPEC, avahi.PROTO_UNSPEC, fqdn, avahi.PROTO_UNSPEC, 0)
	if err != nil {
		log.Fatal("ResolveHostName() failed", err.Error())
	}
	log.Println("ResolveHostName:", hn)

	db, err := server.DomainBrowserNew(avahi.IF_UNSPEC, avahi.PROTO_UNSPEC, "", avahi.DOMAIN_BROWSER_TYPE_BROWSE, 0)
	if err != nil {
		log.Fatal("DomainBrowserNew() failed", err.Error())
	}

	stb, err := server.ServiceTypeBrowserNew(avahi.IF_UNSPEC, avahi.PROTO_UNSPEC, "local", 0)
	if err != nil {
		log.Fatal("ServiceTypeBrowserNew() failed", err.Error())
	}

	sb, err := server.ServiceBrowserNew(avahi.IF_UNSPEC, avahi.PROTO_UNSPEC, "_my-nifty-service._tcp._tcp", "local", 0)
	if err != nil {
		log.Fatal("ServiceBrowserNew() failed", err.Error())
	}

	sr, err := server.ServiceResolverNew(avahi.IF_UNSPEC, avahi.PROTO_UNSPEC, "", "_my-nifty-service._tcp._tcp", "local", avahi.PROTO_UNSPEC, 0)
	if err != nil {
		log.Fatal("ServiceResolverNew() failed", err.Error())
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
							      service.Type, service.Domain, avahi.PROTO_UNSPEC, 0)
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
	"github.com/godbus/dbus"
	"github.com/holoplot/go-avahi"
	"log"
)

func main() {
	conn, err := dbus.SystemBus()
	if err != nil {
		log.Fatal("Cannot get system bus")
	}

	a, err := avahi.ServerNew(conn)
	if err != nil {
		log.Fatal("Avahi new failed")
	}

	eg, err := a.EntryGroupNew()
	if err != nil {
		log.Fatal("EntryGroupNew() failed:", err.Error())
	}

	hostname, err := a.GetHostName()
	if err != nil {
		log.Fatal("GetHostName() failed:", err.Error())
	}

	fqdn, err := a.GetHostNameFqdn()
	if err != nil {
		log.Fatal("GetHostNameFqdn() failed:", err.Error())
	}

	err = eg.AddService(avahi.InterfaceUnspec, avahi.ProtoUnspec, 0, hostname, "_my-nifty-service._tcp", "local", fqdn, 1234, nil)
	if err != nil {
		log.Fatal("AddService() failed:", err.Error())
	}

	err = eg.Commit()
	if err != nil {
		log.Fatal("Commit() failed:", err.Error())
	}

	log.Println("Entry published. Hit ^C to exit.")

	for {
		select {}
	}
}
```

# MIT License

See file `LICENSE` for details.
