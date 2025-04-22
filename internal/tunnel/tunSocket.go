package tunnel

import (
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/google/uuid"
)

type socket struct {
	config SocketConfig
	active bool
	id     string
}

func httpDetect(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("http detect failed, error: %s", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("http detect failed, status code: %d", resp.StatusCode)
	}
	ipaddr, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("http detect failed, error: %s", err)
	}
	return string(ipaddr), nil
}

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
	dynamic := false
	if slices.Contains([]string{AutoHTTPDetect, AutoSTUNDetect}, s.config.PublicAddress) {
		dynamic = true
	}
	return &SocketInfo{
		Address:              address,
		Port:                 port,
		Protocol:             s.config.TransportProtocol,
		Id:                   s.id,
		DynamicPublicAddress: dynamic,
	}, nil
}

func newSocket(config SocketConfig) *socket {
	return &socket{
		config: config,
		active: false,
		id:     uuid.NewString(),
	}
}
