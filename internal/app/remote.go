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
		log.Error("failed to listen local socket for remote app", "localSocket", ra.config.LocalSocket, "remoteApp", ra.config.AppName, "error", err)
		return err
	}
	log.Info("listen local socket for remote app", "localSocket", ra.config.LocalSocket, "remoteSite", ra.config.SiteName, "remoteApp", ra.config.AppName)
	ra.listener = l
	go func() {
		defer func() {
			ra.active.Store(false)
			_ = l.Close()
		}()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				conn, err := l.Accept()
				if err != nil {
					log.Error("failed to accept local socket connection", "localSocket", ra.config.LocalSocket, "remoteSite", ra.config.SiteName, "remoteApp", ra.config.AppName, "error", err)
					return
				}
				log.Info("a new client connection request incoming", "clientAddr", conn.RemoteAddr().String(), "remoteApp", ra.config.AppName)
				go callback(ra.config.SiteName, ra.config.AppName, conn)
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
	log.Info("stop local socket listener for remote app", "remoteSite", ra.config.SiteName, "remoteApp", ra.config.AppName, "localSocket", ra.config.LocalSocket)
	err := ra.listener.Close()
	if err != nil {
		log.Error("failed to close local socket", "localSocket", ra.config.LocalSocket, "remoteSite", ra.config.SiteName, "remoteApp", ra.config.AppName, "error", err)
	}
}

func newRemoteApp(config RemoteAppConfig) *remoteApp {
	return &remoteApp{
		config: config,
		active: atomic.Bool{},
	}
}
