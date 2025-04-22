package tunnel

import (
	"fmt"
	"net/netip"
	"net/url"
	"slices"
)

type HTTPDetector struct {
	URL string `mapstructure:"url"`
}

type SocketConfig struct {
	Mode              SocketMode        `mapstructure:"mode"`
	TransportProtocol TransportProtocol `mapstructure:"transportProtocol"`
	PublicAddress     string            `mapstructure:"publicAddress"`
	PublicePort       uint16            `mapstructure:"publicPort"`
	ListenAddress     string            `mapstructure:"listenAddress"`
	ListenPort        uint16            `mapstructure:"listenPort"`
	HTTPDetector      *HTTPDetector     `mapstructure:"httpDetector"`
}

type Config struct {
	LocalSockets []*SocketConfig `mapstructure:"localSockets"`
}

func CheckAndSetDefaultConfig(config *Config) error {
	if config == nil {
		return nil
	}
	for _, socket := range config.LocalSockets {
		if !slices.Contains([]SocketMode{PortForwarding, DedicatedAddress}, socket.Mode) {
			return fmt.Errorf("unsupported tunnel socket mode: %s", socket.Mode)
		}
		if !slices.Contains([]TransportProtocol{QUIC}, socket.TransportProtocol) {
			return fmt.Errorf("unsupported tunnel transport protocol: %s", socket.TransportProtocol)
		}
		if socket.Mode == DedicatedAddress {
			if socket.PublicAddress == "" && socket.ListenAddress == "" {
				return fmt.Errorf("the 'publicAddress' or 'listenAddress' must be set when the tunnel socket mode set as: %s", DedicatedAddress)
			} else if socket.PublicAddress == "" {
				socket.PublicAddress = socket.ListenAddress
			}
			if socket.ListenPort == 0 && socket.PublicePort == 0 {
				return fmt.Errorf("the 'publicPort' or 'listenPort' must be set when the tunnel socket mode set as: %s", DedicatedAddress)
			} else if socket.PublicePort == 0 {
				socket.PublicePort = socket.ListenPort
			} else if socket.ListenPort == 0 {
				socket.ListenPort = socket.PublicePort
			}
			if socket.ListenPort != socket.PublicePort {
				return fmt.Errorf("the 'publicPort' or 'listenPort' must be equal when the tunnel socket mode set as: %s", DedicatedAddress)
			}
		}
		if socket.Mode == PortForwarding && (socket.PublicAddress == "" || socket.PublicePort == 0 || socket.ListenPort == 0) {
			return fmt.Errorf("the 'publicAddress' and 'publicPort' and 'listenPort' must be set together when the tunnel socket mode set as: %s", PortForwarding)
		}
		switch socket.PublicAddress {
		case AutoHTTPDetect:
			if socket.HTTPDetector == nil {
				return fmt.Errorf("the 'httpDetector' must be set when the 'publicAddress' is set as: %s", AutoHTTPDetect)
			}
			_, err := url.Parse(socket.HTTPDetector.URL)
			if err != nil {
				return fmt.Errorf("the 'httpDetector' url is invalid: %s", err)
			}
			if socket.ListenAddress == "" {
				socket.ListenAddress = "0.0.0.0"
			}
		default:
			addr, err := netip.ParseAddr(socket.PublicAddress)
			if err != nil {
				return fmt.Errorf("the 'publicAddress' is invalid: %s", err)
			}
			if socket.ListenAddress == "" {
				if addr.Is6() {
					socket.ListenAddress = "::"
				} else {
					socket.ListenAddress = "0.0.0.0"
				}
			}
		}
		_, err := netip.ParseAddr(socket.ListenAddress)
		if err != nil {
			return fmt.Errorf("the 'listenAddress' is invalid: %s", err)
		}
	}
	return nil
}
