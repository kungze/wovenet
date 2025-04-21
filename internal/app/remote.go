package app

import (
	"context"
	"io"
	"net"
	"sync/atomic"

	"gihtub.com/kungze/wovenet/internal/logger"
)

type ClientConnectedCallback func(remoteSite string, remoteApp string, conn io.ReadWriteCloser)

type remoteApp struct {
	config   RemoteAppConfig
	stopCh   chan bool
	active   atomic.Bool
	listener net.Listener
}

func (ra *remoteApp) listen(ctx context.Context, callback ClientConnectedCallback) error {
	ra.active.Store(true)
	log := logger.GetDefault()
	networkType := networkType(ra.config.LocalSocket)
	l, err := net.Listen(networkType, ra.config.LocalSocket)
	if err != nil {
		ra.active.Store(false)
		log.Error("failed to listen local socket for remote app", "localSocket", ra.config.LocalSocket, "remoteAppId", ra.config.RemoteAppId, "error", err)
		return err
	}
	log.Info("listen local socket for remote app", "localSocket", ra.config.LocalSocket, "remoteSite", ra.config.SiteName, "remoteAppId", ra.config.RemoteAppId)
	ra.listener = l
	go func() {
		defer func() {
			ra.active.Store(false)
			err := l.Close()
			if err != nil {
				log.Error("failed to close local socket listener for remote app", "localSocket", ra.config.LocalSocket, "remoteSite", ra.config.SiteName, "remoteAppId", ra.config.RemoteAppId, "error", err)
			}
		}()
		for {
			select {
			case <-ra.stopCh:
				return
			case <-ctx.Done():
				return
			default:
				conn, err := l.Accept()
				if err != nil {
					log.Warn("local socket listener encountering error while accepting", "error", err)
					continue
				}
				log.Info("a new client connection request incoming", "clientAddr", conn.RemoteAddr().String(), "remoteAppId", ra.config.RemoteAppId)
				go callback(ra.config.SiteName, ra.config.RemoteAppId, conn)
			}
		}
	}()
	return nil
}

func (ra *remoteApp) Active() bool {
	return ra.active.Load()
}

func (ra *remoteApp) stop() {
	log := logger.GetDefault()
	log.Info("stop local socket listen", "remoteSite", ra.config.SiteName, "appId", ra.config.RemoteAppId, "localSocket", ra.config.LocalSocket)
	err := ra.listener.Close()
	if err != nil {
		log.Error("failed to close local socket", "localSocket", ra.config.LocalSocket, "remoteSite", ra.config.SiteName, "remoteAppId", ra.config.RemoteAppId, "error", err)
	}
	ra.stopCh <- true
}

func newRemoteApp(config RemoteAppConfig) *remoteApp {
	return &remoteApp{
		config: config,
		stopCh: make(chan bool),
		active: atomic.Bool{},
	}
}
