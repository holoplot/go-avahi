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
	Port      int16
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
