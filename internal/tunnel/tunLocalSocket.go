package tunnel

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"slices"

	"github.com/google/uuid"
	"github.com/kungze/wovenet/internal/logger"
)

var supportedDetectMethod = []string{AutoHTTPDetect, AutoSTUNDetect}

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

type tunLocalSocket struct {
	id                         string
	config                     SocketConfig
	active                     bool
	listener                   Listener
	streamCallback             NewStreamCallback
	dataChannelCreatedCallback DataChannelCreatedCallback
	dataChannelDestroyCallback DataChannelDestroyCallback
}

func (s *tunLocalSocket) listenerLoopAccept(ctx context.Context) {
	log := logger.GetDefault()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Accept tunnel connections from remote sites
			conn, err := s.listener.Accept(ctx)
			// TODO(jeffyjf) Whether need to destroy the listener after getting error
			if err != nil {
				log.Error("quic listener encountered an error while accepting a connection", "localAddr", s.listener.Addr().String(), "error", err)
				continue
			}
			// Wait for a control stream to be opened. We accept connection from remote site passively.
			// We can't know the connection from which remote site. So we need a control stream here to
			// tell me the remote site name.
			stream, err := conn.AcceptStream(ctx)
			if err != nil {
				log.Error("QUIC connection encountered an error while accepting a control stream",
					"localAddr", s.listener.Addr().String(),
					"remoteAddr", conn.RemoteAddr().String(), "error", err)
				continue
			}
			buf := make([]byte, 1024)
			n, err := stream.Read(buf)
			if err != nil {
				log.Error("Failed to read handshake data from control stream",
					"localAddr", s.listener.Addr().String(),
					"remoteAddr", conn.RemoteAddr().String(), "error", err)
				_ = stream.Close()
				continue
			}
			len := int(buf[0])
			if n != len+1 {
				_ = stream.Close()
				continue
			}
			// Get remote site name from control stream data
			remoteSite := string(buf[1:n])
			dataChannel := newDataChannel(ctx, conn, remoteSite, s.streamCallback, nil, s.dataChannelDestroyCallback)
			dataChannel.Start()
			go s.dataChannelCreatedCallback(remoteSite, dataChannel)
			log.Info("accept a new remote site connection", "remoteSite", remoteSite, "remoteAddr", conn.RemoteAddr().String())
		}
	}
}

// Start starts the socket and returns a listener
func (s *tunLocalSocket) Start(ctx context.Context) error {
	if s.active {
		return fmt.Errorf("the socket is already active")
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
		return fmt.Errorf("failed to start socket: %s", err)
	}
	s.active = true
	s.listener = listener
	go s.listenerLoopAccept(ctx)
	return nil
}

func (s *tunLocalSocket) IsActive() bool {
	return s.active
}

// GetSocketInfo returns the local tunnel socket information
func (s *tunLocalSocket) GetSocketInfo() (*SocketInfo, error) {
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
		port = s.config.PublicPort
	case AutoSTUNDetect:
		return nil, fmt.Errorf("not implement")
	default:
		address = s.config.PublicAddress
		port = s.config.PublicPort
	}
	return &SocketInfo{
		Address:              address,
		Port:                 port,
		Protocol:             s.config.TransportProtocol,
		DynamicPublicAddress: slices.Contains(supportedDetectMethod, s.config.PublicAddress),
		Id:                   s.id,
	}, nil
}

func newTunLocalSocket(
	config SocketConfig, streamCallback NewStreamCallback,
	dataChannelCreatedCallback DataChannelCreatedCallback,
	dataChannelDestroyCallback DataChannelDestroyCallback) *tunLocalSocket {
	return &tunLocalSocket{
		id:                         uuid.NewString(),
		config:                     config,
		active:                     false,
		streamCallback:             streamCallback,
		dataChannelCreatedCallback: dataChannelCreatedCallback,
		dataChannelDestroyCallback: dataChannelDestroyCallback,
	}
}
