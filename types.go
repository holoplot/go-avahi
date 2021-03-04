package avahi

// Domain ...
type Domain struct {
	Interface int32
	Protocol  int32
	Domain    string
	Flags     uint32
}

// HostName ...
type HostName struct {
	Interface int32
	Protocol  int32
	Name      string
	Aprotocol int32
	Address   string
	Flags     uint32
}

// Address ...
type Address struct {
	Interface int32
	Protocol  int32
	Aprotocol int32
	Address   string
	Name      string
	Flags     uint32
}

// ServiceType ...
type ServiceType struct {
	Interface int32
	Protocol  int32
	Type      string
	Domain    string
	Flags     uint32
}

// Service ...
type Service struct {
	Interface int32
	Protocol  int32
	Name      string
	Type      string
	Domain    string
	Host      string
	Aprotocol int32
	Address   string
	Port      uint16
	Txt       [][]byte
	Flags     uint32
}

// Record ...
type Record struct {
	Interface int32
	Protocol  int32
	Name      string
	Class     int16
	Type      int16
	Rdata     []byte
	Flags     uint32
}

const (
	// ProtoInet - IPv4
	ProtoInet = 0
	// ProtoInet6 - IPv6
	ProtoInet6 = 1
	// ProtoUnspec - Unspecified/all protocol(s)
	ProtoUnspec = -1
)

const (
	// InterfaceUnspec - Unspecified/all interface(s)
	InterfaceUnspec = -1
)

const (
	// PublishUnique - The RRset is intended to be unique
	PublishUnique = 1
	// PublishNoProbe - Though the RRset is intended to be unique no probes shall be sent
	PublishNoProbe = 2
	// PublishNoAnnouce - Do not announce this RR to other hosts
	PublishNoAnnouce = 4
	// PublishAllowMultiple - Allow multiple local records of this type, even if they are intended to be unique
	PublishAllowMultiple = 8
	// PublishNoReverse - don't create a reverse (PTR) entry
	PublishNoReverse = 16
	// PublishNoCookie - do not implicitly add the local service cookie to TXT data
	PublishNoCookie = 32
	// PublishUpdate - Update existing records instead of adding new ones
	PublishUpdate = 64
	// PublishUseWideArea - Register the record using wide area DNS (i.e. unicast DNS update)
	PublishUseWideArea = 128
	// PublishUseMulticast - Register the record using multicast DNS
	PublishUseMulticast = 256
)

const (
	// LookupUseWideArea - Force lookup via wide area DNS
	LookupUseWideArea = 1
	// LookupUseMulticast - Force lookup via multicast DNS
	LookupUseMulticast = 2
	// LookupNoTXT - When doing service resolving, don't lookup TXT record
	LookupNoTXT = 4
	// LookupNoAddreess - When doing service resolving, don't lookup A/AAAA record
	LookupNoAddreess = 8
)

const (
	// LookupResultCached - This response originates from the cache
	LookupResultCached = 1
	// LookupResultWideArea - This response originates from wide area DNS
	LookupResultWideArea = 2
	// LookupResultMulticast  - This response originates from multicast DNS
	LookupResultMulticast = 4
	// LookupResultLocal - This record/service resides on and was announced by the local host. Only available in service and record browsers and only on AVAHI_BROWSER_NEW.
	LookupResultLocal = 8
	// LookupResultOurOwn - This service belongs to the same local client as the browser object. Only available in avahi-client, and only for service browsers and only on AVAHI_BROWSER_NEW.
	LookupResultOurOwn = 16
	// LookupResultStatic - The returned data has been defined statically by some configuration option
	LookupResultStatic = 32
)
