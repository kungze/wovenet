package tunnel

type SocketMode string

const (
	NatTraversal     SocketMode = "nat-traversal"
	DedicatedAddress SocketMode = "dedicated-address"
	PortForwarding   SocketMode = "port-forwarding"
)

type IpProtocol string

const (
	IPv4 IpProtocol = "ipv4"
	IPv6 IpProtocol = "ipv6"
)

type TransportProtocol string

const (
	QUIC TransportProtocol = "quic"
	SCTP TransportProtocol = "SCTP"
)

const (
	STUN string = "stun"
	HTTP string = "http"
)
