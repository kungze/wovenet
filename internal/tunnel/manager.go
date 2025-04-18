package tunnel

import (
	"context"
	"fmt"
	"io"
	"sync"

	"gihtub.com/kungze/wovenet/internal/logger"
)

type TunnelManager struct {
	siteName string
	config   Config
	liteners []Listener
	// map[remoteSite]*tunnel
	tunnels             sync.Map
	remoteSiteGone      RemoteSiteGoneCallback
	remoteSiteConnected RemoteSiteConnectedCallback
	newStream           NewStreamCallback
}

func (tm *TunnelManager) listenerLoopAccept(ctx context.Context, listener Listener) {
	log := logger.GetDefault()
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Accept tunnel connections from remote sites
			conn, err := listener.Accept(ctx)
			// TODO(jeffyjf) Whether need to destroy the listener after getting error
			if err != nil {
				log.Error("QUIC listener encountered an error while accepting a connection", "localAddr", listener.Addr().String(), "error", err)
				continue
			}
			// Wait for a control stream to be opened. We accept connection from remote site passively.
			// We can't know the connection from which remote site. So we need a control stream here to
			// tell me the remote site name.
			stream, err := conn.AcceptStream(ctx)
			if err != nil {
				log.Error("QUIC connection encountered an error while accepting a control stream",
					"localAddr", listener.Addr().String(),
					"remoteAddr", conn.RemoteAddr().String(), "error", err)
				continue
			}
			buf := make([]byte, 1024)
			n, err := stream.Read(buf)
			if err != nil {
				log.Error("Failed to read handshake data from control stream",
					"localAddr", listener.Addr().String(),
					"remoteAddr", conn.RemoteAddr().String(), "error", err)
				stream.Close()
				continue
			}
			len := int(buf[0])
			if n != len+1 {
				stream.Close()
				continue
			}
			// Get remote site name from control stream data
			remoteSite := string(buf[1:n])
			if err := tm.remoteSiteConnected(ctx, remoteSite); err != nil {
				log.Error("Failed to process new remote site", "localAddr", listener.Addr().String(),
					"remoteAddr", conn.RemoteAddr().String(), "remoteSite", remoteSite, "error", err)
				continue
			}
			log.Info("accept a new remote site connection", "remoteSite", remoteSite, "remoteAddr", conn.RemoteAddr().String())
			tun, _ := tm.tunnels.LoadOrStore(remoteSite, newTunnel(tm.newStream, tm.tunnelBroken(remoteSite)))
			tun.(*tunnel).addSlaveConn(ctx, conn)
		}
	}
}

// Start if the tunnel local sockets has configured on the local site,
// The listeners related to these sockets will be try to setup
func (tm *TunnelManager) Start(ctx context.Context) error {
	log := logger.GetDefault()
	for _, config := range tm.config.LocalSockets {
		var listener Listener
		var err error
		switch config.TransportProtocol {
		case QUIC:
			listener, err = newQuicListener(config)
		default:
			log.Warn("unsuported transport protocol for tunnel", "transportProtocol", config.TransportProtocol)
			continue
		}
		if err != nil {
			log.Warn("failed to create quic listener", "localSocket", fmt.Sprintf("%s:%d", config.ListenAddress, config.ListenPort), "error", err)
			continue
		}

		go tm.listenerLoopAccept(ctx, listener)
		tm.liteners = append(tm.liteners, listener)
	}
	if len(tm.config.LocalSockets) != 0 && len(tm.liteners) == 0 {
		return fmt.Errorf("can not create any listener")
	}
	return nil
}

// callback function, which will be called when the all slave connections
// related the loadbalancer are disconnected
func (tm *TunnelManager) tunnelBroken(remoteSite string) tunnelBrokenCallback {
	return func() {
		log := logger.GetDefault()
		tm.tunnels.Delete(remoteSite)
		tm.remoteSiteGone(remoteSite)
		log.Warn("remote site gone", "remoteSite", remoteSite)
	}
}

// OpenNewStream create a new stream in tunnel for local external client and remote app
func (tm *TunnelManager) OpenNewStream(ctx context.Context, siteName string) (io.ReadWriteCloser, error) {
	tun, ok := tm.tunnels.Load(siteName)
	if !ok {
		return nil, fmt.Errorf("can not found tunnl that connect to remote site: %s", siteName)
	}
	return tun.(*tunnel).OpenStream(ctx)
}

// Dial request to establish a new tunnel connection to remote site
func (tm *TunnelManager) Dial(ctx context.Context, remoteSite string, socket SocketInfo) error {
	log := logger.GetDefault()
	log.Info("try to dial remote site tunnel socket listner", "remoteSite", remoteSite, "protocol", socket.Protocol, "remoteAddr", fmt.Sprintf("%s:%d", socket.Address, socket.Port))

	var dialer Dialer
	switch socket.Protocol {
	case QUIC:
		dialer = newQuicDialer(socket)
	case SCTP:
		return fmt.Errorf("unsuported protocol: %s", SCTP)
	default:
		return fmt.Errorf("unsuported protocol: %s", socket.Protocol)
	}

	conn, err := dialer.Dial(ctx)
	if err != nil {
		log.Error("Failed to dial remote site", "remoteSite", remoteSite, "remoteAddr", fmt.Sprintf("%s:%d", socket.Address, socket.Port))
		return err
	}
	log.Info("connect to remote site", "remoteSite", remoteSite, "remoteAddr", fmt.Sprintf("%s:%d", socket.Address, socket.Port))
	// open a control stream, we will tell remote site out site name by this control stream
	stream, err := conn.OpenStream(ctx)
	if err != nil {
		log.Error("failed to open control stream", "remoteSite", remoteSite, "error", err)
		conn.Close()
		return err
	}
	data := []byte(tm.siteName)
	len := byte(len(data))
	n, err := stream.Write(append([]byte{len}, data...))
	if err != nil {
		log.Error("failed to write date to control stream", "remoteSite", remoteSite, "error", err)
		stream.Close()
		conn.Close()
		return err
	}
	if n != int(len)+1 {
		stream.Close()
		conn.Close()
		log.Error("the lenght of data write to control stream is valid", "remoteSite", remoteSite)
		return fmt.Errorf("write data length is not valid")
	}
	if err := tm.remoteSiteConnected(ctx, remoteSite); err != nil {
		log.Error("failed to process remote site", "remoteSite", remoteSite)
		stream.Close()
		conn.Close()
		return err
	}
	tun, _ := tm.tunnels.LoadOrStore(remoteSite, newTunnel(tm.newStream, tm.tunnelBroken(remoteSite)))
	tun.(*tunnel).addSlaveConn(ctx, conn)
	return nil
}

// GetLocalSockets get the local tunnel socket infos, so that the
// remote sites can connect to me by these sockets
func (tm *TunnelManager) GetLocalSockets() ([]SocketInfo, error) {
	socketInfos := []SocketInfo{}
	for _, listener := range tm.liteners {
		info, err := listener.GetSocketInfo()
		if err != nil {
			continue
		}
		socketInfos = append(socketInfos, *info)
	}
	return socketInfos, nil
}

func NewTunnelManager(
	siteName string, config Config, newStream NewStreamCallback,
	remoteSiteConnected RemoteSiteConnectedCallback,
	remoteSiteGone RemoteSiteGoneCallback) (*TunnelManager, error) {
	return &TunnelManager{
		siteName:            siteName,
		config:              config,
		liteners:            []Listener{},
		tunnels:             sync.Map{},
		newStream:           newStream,
		remoteSiteConnected: remoteSiteConnected,
		remoteSiteGone:      remoteSiteGone,
	}, nil
}
