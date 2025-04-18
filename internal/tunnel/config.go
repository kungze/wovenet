package tunnel

import (
	"fmt"
	"net"
	"slices"
)

type SocketConfig struct {
	Mode              SocketMode        `mapstructure:"mode"`
	TransportProtocol TransportProtocol `mapstructure:"transportProtocol"`
	PublicAddress     string            `mapstructure:"publicAddress"`
	PublicePort       int               `mapstructure:"publicPort"`
	ListenAddress     string            `mapstructure:"listenAddress"`
	ListenPort        int               `mapstructure:"listenPort"`
}

type Config struct {
	LocalSockets []SocketConfig `mapstructure:"localSockets"`
}

func CheckAndSetDefaultConfig(config Config) (*Config, error) {
	for _, socket := range config.LocalSockets {
		if !slices.Contains([]SocketMode{PortForwarding, DedicatedAddress}, socket.Mode) {
			return nil, fmt.Errorf("unsupported tunnel socket mode: %s", socket.Mode)
		}
		if !slices.Contains([]TransportProtocol{QUIC}, socket.TransportProtocol) {
			return nil, fmt.Errorf("unsupported tunnel transport protocol: %s", socket.TransportProtocol)
		}
		if socket.Mode == DedicatedAddress {
			if socket.PublicAddress == "" && socket.ListenAddress == "" {
				return nil, fmt.Errorf("the 'publicAddress' or 'listenAddress' must be set when the tunnel socket mode set as: %s", DedicatedAddress)
			} else if socket.PublicAddress == "" {
				socket.PublicAddress = socket.ListenAddress
			}
			if socket.ListenPort == 0 && socket.PublicePort == 0 {
				return nil, fmt.Errorf("the 'publicPort' or 'listenPort' must be set when the tunnel socket mode set as: %s", DedicatedAddress)
			} else if socket.PublicePort == 0 {
				socket.PublicePort = socket.ListenPort
			} else if socket.ListenPort == 0 {
				socket.ListenPort = socket.PublicePort
			}
			if socket.ListenPort != socket.PublicePort {
				return nil, fmt.Errorf("the 'publicPort' or 'listenPort' must be equal when the tunnel socket mode set as: %s", DedicatedAddress)
			}
		}
		if socket.Mode == PortForwarding && (socket.PublicAddress == "" || socket.PublicePort == 0 || socket.ListenPort == 0) {
			return nil, fmt.Errorf("the 'publicAddress' and 'publicPort' and 'listenPort' must be set together when the tunnel socket mode set as: %s", PortForwarding)
		}
		if socket.ListenAddress == "" {
			socket.ListenAddress = "0.0.0.0"
		}
		if net.ParseIP(socket.PublicAddress) == nil {
			return nil, fmt.Errorf("the 'publicAddress' is invalid")
		}
		if net.ParseIP(socket.ListenAddress) == nil {
			return nil, fmt.Errorf("the 'listenAddress' is invalid")
		}
	}
	return &config, nil
}
