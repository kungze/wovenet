package tunnel

import (
	"fmt"
	"io"
	"net/http"

	"gihtub.com/kungze/wovenet/internal/logger"
	"github.com/google/uuid"
)

type socket struct {
	config SocketConfig
	active bool
	id     string
}

func httpDetect(url string) (string, error) {
	log := logger.GetDefault()
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http detect failed, error: %s", err)
	}
	defer resp.Body.Close() //nolint:errcheck
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http detect failed, status code: %d", resp.StatusCode)
	}
	ipaddr, err := io.ReadAll(resp.Body)
	log.Info("get public ip address from http detect", "ipaddr", string(ipaddr))
	if err != nil {
		return "", fmt.Errorf("http detect failed, error: %s", err)
	}
	return string(ipaddr), nil
}

// Start starts the socket and returns a listener
func (s *socket) Start() (Listener, error) {
	if s.active {
		return nil, fmt.Errorf("the socket is already active")
	}
	var err error
	var listener Listener
	switch s.config.TransportProtocol {
	case QUIC:
		listener, err = newQuicListener(&s.config)
	default:
		err = fmt.Errorf("unsupported transport protocol: %s", s.config.TransportProtocol)
	}
	if err != nil {
		s.active = false
		return nil, fmt.Errorf("failed to start socket: %s", err)
	}
	s.active = true
	return listener, nil
}

func (s *socket) Active() bool {
	return s.active
}

// GetSocketInfo returns the local tunnel socket information
func (s *socket) GetSocketInfo() (*SocketInfo, error) {
	if !s.active {
		return nil, fmt.Errorf("the socket is not active")
	}
	var address string
	var port uint16
	switch s.config.PublicAddress {
	case AutoHTTPDetect:
		ipaddr, err := httpDetect(s.config.HTTPDetector.URL)
		if err != nil {
			return nil, err
		}
		address = ipaddr
		port = s.config.PublicePort
	case AutoSTUNDetect:
		return nil, fmt.Errorf("not implement")
	default:
		address = s.config.PublicAddress
		port = s.config.PublicePort
	}
	return &SocketInfo{
		Address:  address,
		Port:     port,
		Protocol: s.config.TransportProtocol,
	}, nil
}

func newSocket(config SocketConfig) *socket {
	return &socket{
		config: config,
		active: false,
		id:     uuid.NewString(),
	}
}
