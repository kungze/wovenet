package app

import (
	"fmt"
	"net"
	"strings"
)

type localApp struct {
	config LocalExposedAppConfig
}

// GetConnection get a connection which connect to the local app
func (la *localApp) GetConnection(socket string) (net.Conn, error) {
	if la.config.Mode == SINGLE {
		socket = la.config.AppSocket
	}
	switch strings.ToLower(la.config.Mode) {
	case SINGLE:
		socket = la.config.AppSocket
	case RANGE:
		s := strings.SplitN(socket, ":", 2)
		if len(s) != 2 {
			return nil, fmt.Errorf("the socket: %s is invalid, the format must be protocol:ipaddr:port", socket)
		}
		addr, port, err := net.SplitHostPort(s[1])
		if err != nil {
			return nil, fmt.Errorf("failed to split socket: %s to addr and port, error: %w", socket, err)
		}
		if !isIPInRange(addr, la.config.AddressRange) || !isPortInRange(port, la.config.PortRange) {
			return nil, fmt.Errorf("can not access the remote site spcified socket: %s, it is not allowed", socket)
		}
	default:
		return nil, fmt.Errorf("app '%s' has invalid mode: %s", la.config.AppName, la.config.Mode)
	}
	s := strings.SplitN(socket, ":", 2)
	if len(s) != 2 {
		return nil, fmt.Errorf("the appSocket: %s is invalid, the format must be protocol:ipaddr:port", la.config.AppSocket)
	}
	conn, err := net.Dial(strings.ToLower(s[0]), s[1])
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func newLocalApp(config LocalExposedAppConfig) *localApp {
	return &localApp{config: config}
}
