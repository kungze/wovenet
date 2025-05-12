package tunnel

import (
	"context"
	"fmt"
	"io"
	"sync"

	"github.com/kungze/wovenet/internal/logger"
)

type TunnelManager struct {
	siteName               string
	config                 *Config
	localSockets           map[string]*tunLocalSocket
	remoteSockets          sync.Map // map[socketInfo.Id]*tunRemoteSocket
	tunnels                map[string]*tunnel
	remoteSiteDisconnected RemoteSiteDisconnectedCallback
	remoteSiteConnected    RemoteSiteConnectedCallback
	newStream              NewStreamCallback
	requestNewSocketInfo   RequestNewRemoteSocketInfo
	tunMux                 sync.Mutex
}

// Start if the tunnel local sockets has configured on the local site,
// The listeners related to these sockets will be try to setup
func (tm *TunnelManager) Start(ctx context.Context) error {
	log := logger.GetDefault()
	if tm.config == nil || len(tm.config.LocalSockets) == 0 {
		log.Info("no local socket configured, skip to start local tunnel listeners")
		return nil
	}
	for _, config := range tm.config.LocalSockets {
		socket := newTunLocalSocket(*config, tm.newStream, tm.onDataChannelCreated, tm.onDataChannelDestroyed)
		tm.localSockets[socket.id] = socket
		err := socket.Start(ctx)
		if err != nil {
			log.Error("failed to start socket listener", "socket", config, "error", err)
			continue
		}
		tm.localSockets[socket.id] = socket
	}
	if len(tm.localSockets) == 0 {
		log.Warn("no local tunnel socket listener is started")
	}
	return nil
}

func (tm *TunnelManager) onDataChannelCreated(remoteSite string, dc *dataChannel) {
	log := logger.GetDefault()
	tm.tunMux.Lock()
	defer tm.tunMux.Unlock()
	go tm.remoteSiteConnected(context.Background(), remoteSite)
	tunnel, ok := tm.tunnels[remoteSite]
	if !ok {
		tunnel = newTunnel(remoteSite, tm.tunnelBroken(remoteSite))
		tm.tunnels[remoteSite] = tunnel
	}
	log.Info("add a slave data channel to tunnel", "remoteSite", remoteSite, "channelId", dc.GetId())
	tunnel.AddSlaveDataChannel(dc)
}

func (tm *TunnelManager) onDataChannelDestroyed(remoteSite string, channelId string) {
	log := logger.GetDefault()
	tm.tunMux.Lock()
	defer tm.tunMux.Unlock()
	tunnel, ok := tm.tunnels[remoteSite]
	if !ok {
		return
	}
	log.Info("remove a slave data channel from tunnel", "remoteSite", remoteSite, "channelId", channelId)
	tunnel.DeleteSlaveDataChannel(channelId)
}

// callback function, which will be called when the all slave connections
// belong to the tunnel which connect to the remoteSite are disconnected
func (tm *TunnelManager) tunnelBroken(remoteSite string) tunnelBrokenCallback {
	return func() {
		log := logger.GetDefault()
		log.Warn("the tunnel to remote site is broken", "remoteSite", remoteSite)
		tm.tunMux.Lock()
		defer tm.tunMux.Unlock()
		delete(tm.tunnels, remoteSite)
		tm.remoteSiteDisconnected(remoteSite)
	}
}

// OpenNewStream create a new stream in tunnel for local external client and remote app
func (tm *TunnelManager) OpenNewStream(ctx context.Context, siteName string) (io.ReadWriteCloser, error) {
	tun, ok := tm.tunnels[siteName]
	if !ok {
		return nil, fmt.Errorf("can not found tunnel that connect to remote site: %s", siteName)
	}
	return tun.OpenStream(ctx)
}

// Dial request to establish a new tunnel connection to remote site
func (tm *TunnelManager) AddRemoteSocket(ctx context.Context, remoteSite string, socket SocketInfo) error {
	log := logger.GetDefault()

	// check if the socket is already connected
	// if the socket is already connected, we will not dial it again
	var remoteSocket *tunRemoteSocket
	value, ok := tm.remoteSockets.Load(socket.Id)
	if ok {
		// if the remote socket is already added, we will update the socket info
		// and try to open a new data channel
		remoteSocket, _ = value.(*tunRemoteSocket)
		remoteSocket.SocketInfo = socket
	} else {
		remoteSocket = newTunRemoteSocket(ctx, socket, tm.siteName, remoteSite, tm.requestNewSocketInfo, tm.newStream, tm.onDataChannelCreated, tm.onDataChannelDestroyed)
		tm.remoteSockets.Store(socket.Id, remoteSocket)
	}

	err := remoteSocket.OpenDataChannel(ctx)
	if err != nil {
		log.Error("failed to connect remote socket", "remoteSite", remoteSite, "socketInfo", socket, "error", err)
		return err
	}

	go tm.remoteSiteConnected(ctx, remoteSite)

	return nil
}

func (tm *TunnelManager) DelRemoteSocket(ctx context.Context, remoteSite string, socket SocketInfo) {
	log := logger.GetDefault()
	value, ok := tm.remoteSockets.Load(socket.Id)
	if !ok {
		log.Warn("the remote socket is not found", "remoteSite", remoteSite, "socketInfo", socket)
		return
	}
	remoteSocket, _ := value.(*tunRemoteSocket)
	if remoteSocket != nil {
		remoteSocket.Destroy()
		log.Info("destroy remote socket", "remoteSite", remoteSite, "socketInfo", socket)
	}
	tm.remoteSockets.Delete(socket.Id)
}

// GetLocalSocketInfos get the local tunnel socket infos, so that the
// remote sites can connect to me by these sockets
func (tm *TunnelManager) GetLocalSocketInfos() []SocketInfo {
	log := logger.GetDefault()
	socketInfos := []SocketInfo{}
	for _, socket := range tm.localSockets {
		if !socket.IsActive() {
			log.Warn("the local tunnel socket is not active", "tunnelSocket", socket.config)
			continue
		}
		info, err := socket.GetSocketInfo()
		if err != nil {
			log.Warn("failed to get local tunnel socket info", "error", err, "tunnelSocket", socket.config)
			continue
		}
		socketInfos = append(socketInfos, *info)
	}
	return socketInfos
}

func (tm *TunnelManager) GetLocalSocketInfoById(id string) (*SocketInfo, error) {
	socket, ok := tm.localSockets[id]
	if !ok {
		return nil, fmt.Errorf("can not found socket by id: %s", id)
	}
	if !socket.IsActive() {
		return nil, fmt.Errorf("the socket is not active")
	}
	info, err := socket.GetSocketInfo()
	if err != nil {
		return nil, err
	}
	return info, nil
}

func NewTunnelManager(
	siteName string, config *Config, newStream NewStreamCallback,
	remoteSiteConnected RemoteSiteConnectedCallback,
	remoteSiteDisconnected RemoteSiteDisconnectedCallback,
	requestNewSocketInfo RequestNewRemoteSocketInfo) (*TunnelManager, error) {
	return &TunnelManager{
		siteName:               siteName,
		config:                 config,
		tunnels:                map[string]*tunnel{},
		newStream:              newStream,
		remoteSiteConnected:    remoteSiteConnected,
		remoteSiteDisconnected: remoteSiteDisconnected,
		requestNewSocketInfo:   requestNewSocketInfo,
		localSockets:           make(map[string]*tunLocalSocket),
		remoteSockets:          sync.Map{},
		tunMux:                 sync.Mutex{},
	}, nil
}
