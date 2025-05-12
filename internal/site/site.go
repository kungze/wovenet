package site

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/kungze/wovenet/internal/app"
	"github.com/kungze/wovenet/internal/crypto"
	"github.com/kungze/wovenet/internal/logger"
	"github.com/kungze/wovenet/internal/message"
	"github.com/kungze/wovenet/internal/tunnel"
	"github.com/spf13/viper"
)

// AppInfo it will be packaged as the handshake data in the first packet of each app stream.
// It informs the remote site which app the external client from the local site wants to access
// via the app stream
type AppInfo struct {
	Name   string `json:"name"`
	Socket string `json:"socket"`
}

type Site struct {
	msgClient     message.MessageClient
	siteName      string
	crypto        *crypto.Crypto
	tunnelManager *tunnel.TunnelManager
	appManager    *app.AppManager
	remoteSites   sync.Map
	ctx           context.Context
}

// Start start a local site
// 1. Startup some listeners on local tunnel sockets
// 2. Announce this site's base information by publish message
func (s *Site) Start(ctx context.Context) error {
	log := logger.GetDefault()
	if err := s.tunnelManager.Start(ctx); err != nil {
		log.Error("failed to start tunnel manager", "error", err)
		return nil
	}

	sockets := s.tunnelManager.GetLocalSocketInfos()
	exposedApps := s.appManager.GetExposedApps()
	data := siteInfo{
		TunnelListenerSockets: sockets,
		ExposedApps:           exposedApps,
	}
	// Announce a new site online with the site's base info
	// The first message is to request exchange the base information with each other sites
	err := s.msgClient.BroadcastMessage(ctx, message.ExchangeInfoRequest, data)
	if err != nil {
		log.Error("failed to publish message", "error", err)
		return err
	}
	return nil
}

// onExchangeInfoMessage the callback function for message channel receive exchange
// information request or response. It usually means that a new remote site online
func (s *Site) onExchangeInfoMessage(payload *message.Payload) (any, message.MessageKind, error) {
	log := logger.GetDefault()

	// Decode the message payload data and get the remote site's base information
	info := siteInfo{}
	err := mapstructure.Decode(payload.Data, &info)
	if err != nil {
		log.Error("failed to decode message payload", "error", err)
		return nil, "", err
	}
	log.Info("receive remote site base info", "remoteSite", payload.SiteName)
	value, ok := s.remoteSites.Load(payload.SiteName)
	if ok {
		oldSiteInfo := value.(*siteInfo)
		for _, socket := range oldSiteInfo.TunnelListenerSockets {
			s.tunnelManager.DelRemoteSocket(s.ctx, payload.SiteName, socket)
		}
	}
	s.remoteSites.Store(payload.SiteName, &info)
	// Try to connect to the new remote site
	for _, socket := range info.TunnelListenerSockets {
		err = s.tunnelManager.AddRemoteSocket(s.ctx, payload.SiteName, socket)
		if err != nil {
			log.Warn("failed to establish tunnel with remote site", "remoteSite", payload.SiteName, "error", err)
		}
	}

	// Respond the request message with our base information
	if payload.Kind == message.ExchangeInfoRequest {
		sockets := s.tunnelManager.GetLocalSocketInfos()
		exposedApps := s.appManager.GetExposedApps()

		return &siteInfo{
			TunnelListenerSockets: sockets,
			ExposedApps:           exposedApps,
		}, message.ExchangeInfoResponse, nil
	}

	return nil, "", nil
}

// onNewSocketInfoRequest the callback function for message channel receive a request
// for a new socket info. The request is triggered by a remote site when the remote site
// encounter a connection error with the local site's socket. the local site will respond
// the request with a new socket info which contains a new public address.
func (s *Site) onNewSocketInfoRequest(payload *message.Payload) (any, message.MessageKind, error) {
	log := logger.GetDefault()
	// Decode the message payload data and get the remote site's base information
	request := tunnel.SocketInfoRequest{}
	err := mapstructure.Decode(payload.Data, &request)
	if err != nil {
		log.Error("failed to decode message payload", "error", err)
		return nil, "", err
	}
	// Get the local socket info by id
	socketInfo, err := s.tunnelManager.GetLocalSocketInfoById(request.Id)
	if err != nil {
		log.Error("failed to get local socket info", "error", err, "socketId", request.Id)
		return nil, "", err
	}
	return socketInfo, message.NewSocketInfoResponse, nil
}

// onNewSocketInfoResponse the callback function for message channel receive a response
// for a new socket info(maybe contains a new public address).
func (s *Site) onNewSocketInfoResponse(payload *message.Payload) (any, message.MessageKind, error) {
	log := logger.GetDefault()
	// Decode the message payload data and get the remote site's base information
	info := tunnel.SocketInfo{}
	err := mapstructure.Decode(payload.Data, &info)
	if err != nil {
		log.Error("failed to decode message payload", "error", err)
		return nil, "", err
	}
	err = s.tunnelManager.AddRemoteSocket(s.ctx, payload.SiteName, info)
	if err != nil {
		log.Error("failed to connect to remote site", "remoteSite", payload.SiteName, "socketINfo", info, "error", err)
		return nil, "", err
	}
	return nil, "", nil
}

// requestNewRemoteSocketInfo request a new socket info from remote site, it will be
// called when the local site encounter a connection error with the remote site's socket
func (s *Site) requestNewRemoteSocketInfo(remoteSite string, socketId string) error {
	log := logger.GetDefault()
	// Get the remote site info
	_, ok := s.remoteSites.Load(remoteSite)
	if !ok {
		log.Error("can not found remote site info", "remoteSite", remoteSite)
		return fmt.Errorf("can not found remote site info")
	}
	// Send a request message to the remote site
	err := s.msgClient.UnicastMessage(s.ctx, remoteSite, message.NewSocketInfoRequest, tunnel.SocketInfoRequest{Id: socketId})
	if err != nil {
		log.Error("failed to send new socket info request", "remoteSite", remoteSite, "error", err)
		return err
	}
	return nil
}

// onRemoteSiteDisconnected callback function, which will be called when
// a remote site is disconnected (the tunnel to the remoteSite is broken)
func (s *Site) onRemoteSiteDisconnected(remoteSite string) {
	log := logger.GetDefault()
	log.Info("remote site is disconnected", "remoteSite", remoteSite)
	s.appManager.ProcessRemoteSiteGone(remoteSite)
}

// onRemoteSiteConnected callback function, which will be called when a new remote site
// connects to the local site successfully
func (s *Site) onRemoteSiteConnected(ctx context.Context, remoteSite string) {
	log := logger.GetDefault()
	info, ok := s.remoteSites.Load(remoteSite)
	if !ok {
		// Because the remote site info is received through message channel, it might arrive later
		log.Warn("can not found remote site info, wait 5 second and then check again", "remoteSite", remoteSite)
		<-time.NewTicker(5 * time.Second).C
		info, ok = s.remoteSites.Load(remoteSite)
		if !ok {
			log.Error("can not found remote site info", "remoteSite", remoteSite)
			return
		}
	}
	s.appManager.ProcessNewRemoteSite(ctx, remoteSite, info.(*siteInfo).ExposedApps, s.onNewClientConnection)
}

// onNewClientConnection callback function, which will be called when an external client connects to the
// local socket which is listened for remote app
func (s *Site) onNewClientConnection(remoteSite string, appName string, appSocket string, conn io.ReadWriteCloser) {
	defer conn.Close() //nolint:errcheck
	log := logger.GetDefault()

	// Open a new stream in the tunnel which link the local site and the remote site
	// which the remote app is located in
	log.Info("try to open a new stream in tunnel to connect to remote app", "remoteSite", remoteSite, "remoteApp", appName)
	stream, err := s.tunnelManager.OpenNewStream(s.ctx, remoteSite)
	if err != nil {
		log.Error("failed to open a new stream", "remoteSite", remoteSite, "error", err)
		return
	}
	defer stream.Close() //nolint:errcheck

	handShake, err := json.Marshal(AppInfo{Name: appName, Socket: appSocket})
	if err != nil {
		log.Error("failed to marshal app info", "error", err)
		return
	}

	encrBuf, err := s.crypto.Encrypt(handShake)
	if err != nil {
		log.Error("failed to encrypt handshake data", "error", err)
		return
	}

	// Prepare the handshake data, to tell remote site we want to connect to which app
	data := []byte(encrBuf)
	dataLen := make([]byte, 2)
	binary.LittleEndian.PutUint16(dataLen, uint16(len(data)))
	log.Info("try to write handshake data to app stream", "remoteSite", remoteSite, "remoteApp", appName)
	n, err := stream.Write(append(dataLen, data...))
	if err != nil {
		log.Error("failed to write handshake data to app stream",
			"remoteSite", remoteSite, "remoteApp", appName, "error", err)
		return
	}
	if n != len(data)+2 {
		log.Error("the length of handshake data is valid", "expectedLen", len(data)+2, "actualLen", n)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer func() {
			log.Warn("the coroutine which copy data from tunnel stream to local client exit",
				"remoteSite", remoteSite, "remoteApp", appName)
			_ = conn.Close()
			_ = stream.Close()
		}()
		log.Info("start to copy data from tunnel stream to local client",
			"remoteSite", remoteSite, "remoteApp", appName)
		defer wg.Done()
		_, err := io.Copy(conn, stream)
		if err != nil {
			log.Error("failed to copy data from tunnel stream to local client",
				"remoteSite", remoteSite, "remoteApp", appName, "error", err)
		}
	}()
	go func() {
		defer func() {
			log.Warn("the coroutine which copy data from local client to tunnel stream exit",
				"remoteSite", remoteSite, "remoteApp", appName)
			_ = conn.Close()
			_ = stream.Close()
		}()
		log.Info("start to copy data from local client to tunnel stream",
			"remoteSite", remoteSite, "remoteApp", appName)
		defer wg.Done()
		_, err := io.Copy(stream, conn)
		if err != nil {
			log.Error("failed to copy data from local client to tunnel stream",
				"remoteSite", remoteSite, "remoteApp", appName, "error", err)
		}
	}()
	wg.Wait()
}

// onNewStream call function, which will be called when a new stream was accepted
// it means that a external client from remote site want to connect to the local
// site's local exposed app
func (s *Site) onNewStream(stream tunnel.Stream) {
	log := logger.GetDefault()
	log.Info("a new stream was accepted")
	defer stream.Close() //nolint:errcheck
	// Read handshake data, the handshake data indicates the remote client
	// want to access which app
	buff := make([]byte, 1024)
	n, err := stream.Read(buff)
	if err != nil {
		log.Error("failed to read handshake data from tunnel stream", "error", err)
		return
	}
	if n < 2 {
		log.Error("failed to read handshake data, the data length is too short")
		return
	}
	// Get the handshake data length
	dataLen := binary.LittleEndian.Uint16(buff[:2])
	if n < int(dataLen+2) {
		log.Error("the handshake data length is valid", "expectedLen", dataLen+2, "actualLen", n)
		return
	}
	// Get app name from handshake data
	decBuf, err := s.crypto.Decrypt(string(buff[2 : dataLen+2]))
	if err != nil {
		log.Error("failed to decrypt handshake data")
		return
	}

	appInfo := AppInfo{}
	err = json.Unmarshal(decBuf, &appInfo)
	if err != nil {
		log.Error("failed to unmarshal app info", "error", err)
		return
	}
	if appInfo.Socket != "" {
		log.Info("the remote site has specified app socket", "localApp", appInfo.Name, "socket", appInfo.Socket)
	}

	log.Info("try to connect to local app", "localApp", appInfo.Name)
	conn, err := s.appManager.ConnectToLocalApp(appInfo.Name, appInfo.Socket)
	if err != nil {
		log.Error("failed to connect to local app", "localApp", appInfo.Name, "error", err)
		return
	}
	defer conn.Close() //nolint:errcheck
	if n > int(dataLen+2) {
		_, err := conn.Write(buff[dataLen+2 : n])
		if err != nil {
			log.Error("failed to write data to local app", "localApp", appInfo.Name, "error", err)
			return
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer func() {
			log.Warn("the coroutine which copy data from tunnel stream to local app exit", "localApp", appInfo.Name)
			_ = conn.Close()
			_ = stream.Close()
		}()
		log.Info("start to copy data from tunnel stream to local app", "localApp", appInfo.Name)
		defer wg.Done()
		_, err := io.Copy(conn, stream)
		if err != nil {
			log.Error("failed to copy data from tunnel stream to local app", "localApp", appInfo.Name)
		}
	}()
	go func() {
		defer func() {
			log.Warn("the coroutine which copy data from local app to tunnel stream exit", "localApp", appInfo.Name)
			_ = conn.Close()
			_ = stream.Close()
		}()
		log.Info("start to copy data from local app to tunnel stream", "localApp", appInfo.Name)
		defer wg.Done()
		_, err := io.Copy(stream, conn)
		if err != nil {
			log.Error("failed to copy data from local app to tunnel stream", "localApp", appInfo.Name)
		}
	}()
	wg.Wait()
}

func NewSite(ctx context.Context) (*Site, error) {
	log := logger.GetDefault()
	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Error("failed to unmarshal the config into a struct", "error", err)
		return nil, err
	}

	err = CheckAndSetDefaultConfig(&config)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	crypto, err := crypto.NewCrypto([]byte(config.Crypto.Key))
	if err != nil {
		log.Error("failed to create crypto", "error", err)
		return nil, err
	}

	log.Info("new local site", "siteName", config.SiteName)
	site := &Site{
		siteName:    config.SiteName,
		crypto:      crypto,
		remoteSites: sync.Map{},
		ctx:         ctx,
	}

	am, err := app.NewAppManager(ctx, config.LocalExposedApps, config.RemoteApps)
	if err != nil {
		log.Error("failed to create app manager", "error", err)
		return nil, err
	}
	site.appManager = am
	tm, err := tunnel.NewTunnelManager(
		config.SiteName, config.Tunnel, site.onNewStream,
		site.onRemoteSiteConnected, site.onRemoteSiteDisconnected,
		site.requestNewRemoteSocketInfo)
	if err != nil {
		log.Error("failed to create tunnel manager", "error", err)
		return nil, err
	}
	site.tunnelManager = tm

	msgClient, err := message.NewMessageClient(ctx, *config.MessageChannel, *config.Crypto, site.siteName)
	if err != nil {
		log.Error("failed to create message client", "error", err)
		return nil, err
	}
	msgClient.RegisterHandler(message.ExchangeInfoRequest, site.onExchangeInfoMessage)
	msgClient.RegisterHandler(message.ExchangeInfoResponse, site.onExchangeInfoMessage)
	msgClient.RegisterHandler(message.NewSocketInfoRequest, site.onNewSocketInfoRequest)
	msgClient.RegisterHandler(message.NewSocketInfoResponse, site.onNewSocketInfoResponse)
	site.msgClient = msgClient

	return site, nil
}
