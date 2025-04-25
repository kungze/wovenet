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
func (la *localApp) GetConnection() (net.Conn, error) {
	s := strings.SplitN(la.config.AppSocket, ":", 2)
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
