package app

import (
	"net"
)

type localApp struct {
	config LocalExposedAppConfig
}

// GetConnection get a connection which connect to the local app
func (la *localApp) GetConnection() (net.Conn, error) {
	network := networkType(la.config.Socket)
	conn, err := net.Dial(network, la.config.Socket)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func newLocalApp(config LocalExposedAppConfig) *localApp {
	return &localApp{config: config}
}
