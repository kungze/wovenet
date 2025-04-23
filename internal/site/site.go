package site

import (
	"context"
	"fmt"
	"io"
	"sync"
	"time"

	"gihtub.com/kungze/wovenet/internal/app"
	"gihtub.com/kungze/wovenet/internal/logger"
	"gihtub.com/kungze/wovenet/internal/message"
	"gihtub.com/kungze/wovenet/internal/tunnel"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

type Site struct {
	msgClient     message.MessageClient
	siteName      string
	tunnelManager *tunnel.TunnelManager
	appManager    *app.AppManager
	remoteSites   sync.Map
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

	// Get the local listenr sockets for others sites to connect to establish tunnels
	sockets, err := s.tunnelManager.GetLocalSocketInfos()
	if err != nil {
		log.Error("failed to get tunnel local sockets", "error", err)
		return err
	}
	exposedApps := s.appManager.GetExposedApps()
	data := siteInfo{
		TunnelListenerSockets: sockets,
		ExposedApps:           exposedApps,
	}
	// Announce a new site online with the site's base info
	// The first message is to request exchange the base information with each other sites
	err = s.msgClient.BroadcastMessage(context.Background(), message.ExchangeInfoRequest, data)
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
	s.remoteSites.Store(payload.SiteName, &info)
	// Try to connect to the new remote site
	for _, socket := range info.TunnelListenerSockets {
		err = s.tunnelManager.Dial(context.Background(), payload.SiteName, socket)
		if err != nil {
			log.Error("failed to establish tunnel with remote site", "remoteSite", payload.SiteName, "error", err)
			return nil, "", err
		}
	}

	// Respond the request message with our base information
	if payload.Kind == message.ExchangeInfoRequest {
		sockets, err := s.tunnelManager.GetLocalSocketInfos()
		if err != nil {
			log.Error("failed to get tunnel local sockets", "error", err)
			return nil, "", err
		}
		exposedApps := s.appManager.GetExposedApps()

		return &siteInfo{
			TunnelListenerSockets: sockets,
			ExposedApps:           exposedApps,
		}, message.ExchangeInfoResponse, nil
	}

	return nil, "", nil
}

// onRemoteSiteGone callback function, which will be called when
// a remote site is disconnected (all connections to the remote site are unusable)
func (s *Site) onRemoteSiteGone(remoteSite string) {
	s.appManager.ProcessRemoteSiteGone(remoteSite)
	s.remoteSites.Delete(remoteSite)
}

// onRemoteSiteConnected callback function, which will be called when a new remote site
// connects to the local site successfully
func (s *Site) onRemoteSiteConnected(ctx context.Context, remoteSite string) error {
	log := logger.GetDefault()
	info, ok := s.remoteSites.Load(remoteSite)
	if !ok {
		// Because the remote site info is received through message channel, it might arrive later
		log.Warn("can not found remote site info, wait 5 second and then check again", "remoteSite", remoteSite)
		<-time.NewTicker(5 * time.Second).C
		info, ok = s.remoteSites.Load(remoteSite)
		if !ok {
			return fmt.Errorf("can not get remote site: %s info", remoteSite)
		}
	}
	return s.appManager.ProcessNewRemoteSite(ctx, remoteSite, info.(*siteInfo).ExposedApps, s.onNewClientConnection)
}

// onNewClientConnection callback function, which will be called when an external client connects to the
// local socket which is listened for remote app
func (s *Site) onNewClientConnection(remoteSite string, remoteApp string, conn io.ReadWriteCloser) {
	log := logger.GetDefault()

	defer func() {
		if err := conn.Close(); err != nil {
			log.Error("failed to close connection", "remoteSite", remoteSite, "remoteApp", remoteApp, "error", err)
		}
	}()

	// Open a new strem in the tunnel which link the local site and the remote site
	// which the remote app is located in
	stream, err := s.tunnelManager.OpenNewStream(context.Background(), remoteSite)
	if err != nil {
		log.Error("failed to open a new stream", "remoteSite", remoteSite, "error", err)
		return
	}
	defer stream.Close() //nolint:errcheck

	// Prepare the handshake data, to tell remote site we want to connect to which app
	data := []byte(remoteApp)
	len := byte(len(data))
	n, err := stream.Write(append([]byte{len}, data...))
	if err != nil {
		log.Error("failed to write handshake data to app stream",
			"remoteSite", remoteSite, "remoteApp", remoteApp, "error", err)
		return
	}
	if n != int(len)+1 {
		log.Error("the lenght of handshake data is valid", "expectedLen", int(len)+1, "actualLen", n)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer conn.Close()   //nolint:errcheck
		defer stream.Close() //nolint:errcheck
		_, err := io.Copy(conn, stream)
		if err != nil {
			log.Error("failed to copy data from tunnel stream to local client",
				"remoteSite", remoteSite, "remoteApp", remoteApp, "error", err)
		}
		log.Warn("the coroutine which copy data from tunnel stream to local client exit",
			"remoteSite", remoteSite, "remoteApp", remoteApp)
	}()
	go func() {
		defer wg.Done()
		defer conn.Close()   //nolint:errcheck
		defer stream.Close() //nolint:errcheck
		_, err := io.Copy(stream, conn)
		if err != nil {
			log.Error("failed to copy data from local client to tunnel stream",
				"remoteSite", remoteSite, "remoteApp", remoteApp, "error", err)
		}
		log.Warn("the coroutine which copy data from local client to tunnel stream exit",
			"remoteSite", remoteSite, "remoteApp", remoteApp)
	}()
	wg.Wait()
}

// onNewStream call function, which will be called when a new stream was accepted
// it means that a external client from remote site want to connect to our local app
func (s *Site) onNewStream(stream tunnel.Stream) {
	log := logger.GetDefault()
	defer stream.Close() //nolint:errcheck
	// Read handshake data, the handshake data indicates the remote client
	// want to access which app
	buff := make([]byte, 1024)
	n, err := stream.Read(buff)
	if err != nil {
		log.Error("failed to read handshake data from tunnl stream", "error", err)
		return
	}
	len := int(buff[0])
	if n < len+1 {
		log.Error("the handshake data lenght is valid", "expectedLen", len+1, "accutalLen", n)
		return
	}
	// Get app id from handshake data
	appId := string(buff[1 : len+1])
	log.Info("try to connect to local app", "appId", appId)
	conn, err := s.appManager.ConnectToLocalApp(appId)
	if err != nil {
		log.Error("failed to connect to local app", "appId", appId)
	}
	defer conn.Close() //nolint:errcheck
	if n > len+1 {
		_, err := conn.Write(buff[len+1 : n])
		if err != nil {
			log.Error("failed to write data to local app", "appId", appId)
			return
		}
	}
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer wg.Done()
		defer conn.Close()   //nolint:errcheck
		defer stream.Close() //nolint:errcheck
		_, err := io.Copy(conn, stream)
		if err != nil {
			log.Error("failed to copy data from tunnel stream to local app", "appId", appId)
		}
		log.Warn("the coroutine which copy data from tunnel stream to local app exit", "appId", appId)
	}()
	go func() {
		defer wg.Done()
		defer conn.Close()   //nolint:errcheck
		defer stream.Close() //nolint:errcheck
		_, err := io.Copy(stream, conn)
		if err != nil {
			log.Error("failed to copy data from local app to tunnel stream", "appId", appId)
		}
		log.Warn("the coroutine which copy data from local app to tunnel stream exit", "appId", appId)
	}()
	wg.Wait()
}

func NewSite(ctx context.Context) (*Site, error) {
	log := logger.GetDefault()
	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Error("failed to unmarshals the config into a struct", "error", err)
		return nil, err
	}

	err = CheckAndSetDefaultConfig(&config)
	if err != nil {
		return nil, fmt.Errorf("config error: %w", err)
	}

	log.Info("new local site", "siteName", config.SiteName)
	site := &Site{
		siteName:    config.SiteName,
		remoteSites: sync.Map{},
	}

	am, err := app.NewAppManager(ctx, config.LocalExposedApps, config.RemoteApps)
	if err != nil {
		log.Error("failed to create app manager", "error", err)
		return nil, err
	}
	site.appManager = am
	tm, err := tunnel.NewTunnelManager(config.SiteName, config.Tunnel, site.onNewStream, site.onRemoteSiteConnected, site.onRemoteSiteGone)
	if err != nil {
		log.Error("failed to create tunnel manager", "error", err)
		return nil, err
	}
	site.tunnelManager = tm

	msgClient, err := message.NewMessageClient(ctx, *config.MessageChannel, site.siteName)
	if err != nil {
		log.Error("failed to create message client", "error", err)
		return nil, err
	}
	msgClient.RegisterHandler(message.ExchangeInfoRequest, site.onExchangeInfoMessage)
	msgClient.RegisterHandler(message.ExchangeInfoResponse, site.onExchangeInfoMessage)
	site.msgClient = msgClient

	return site, nil
}
