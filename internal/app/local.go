package app

import (
	"fmt"
	"net"
	"strings"

	"github.com/kungze/wovenet/internal/tunnel"
)

type localApp struct {
	config LocalExposedAppConfig
}

// StartDataConverter start a converter to transfer data between tunnel stream and local app
func (la *localApp) StartDataConverter(stream tunnel.Stream, socket string, remainingData []byte) error {
	if la.config.Mode == SINGLE {
		socket = la.config.AppSocket
	}
	switch strings.ToLower(la.config.Mode) {
	case SINGLE:
		socket = la.config.AppSocket
	case RANGE:
		s := strings.SplitN(socket, ":", 2)
		if len(s) != 2 {
			return fmt.Errorf("the socket: %s is invalid, the format must be protocol:ipaddr:port", socket)
		}
		addr, port, err := net.SplitHostPort(s[1])
		if err != nil {
			return fmt.Errorf("failed to split socket: %s to addr and port, error: %w", socket, err)
		}
		if !isIPInRange(addr, la.config.AddressRange) || !isPortInRange(port, la.config.PortRange) {
			return fmt.Errorf("can not access the remote site specified socket: %s, it is not allowed", socket)
		}
	default:
		return fmt.Errorf("app '%s' has invalid mode: %s", la.config.AppName, la.config.Mode)
	}
	s := strings.SplitN(socket, ":", 2)
	if len(s) != 2 {
		return fmt.Errorf("the appSocket: %s is invalid, the format must be protocol:ipaddr:port", la.config.AppSocket)
	}
	conn, err := net.Dial(strings.ToLower(s[0]), s[1])
	if err != nil {
		return err
	}
	if len(remainingData) > 0 {
		n, err := conn.Write(remainingData)
		if err != nil {
			return fmt.Errorf("failed to write data to local app: %s, error: %w", la.config.AppName, err)
		}
		if n < len(remainingData) {
			return fmt.Errorf("can not write all remaining data to local app: %s", la.config.AppName)
		}
	}
	go la.startConverter(stream, conn)
	return nil
}

func (la *localApp) startConverter(stream tunnel.Stream, conn net.Conn) {
	c := converter{
		conn:    conn,
		stream:  stream,
		appType: localAppType,
		appName: la.config.AppName,
	}
	c.Start()
}

func newLocalApp(config LocalExposedAppConfig) *localApp {
	return &localApp{config: config}
}
