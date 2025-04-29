package site

import (
	"github.com/kungze/wovenet/internal/app"
	"github.com/kungze/wovenet/internal/tunnel"
)

type siteInfo struct {
	TunnelListenerSockets []tunnel.SocketInfo   `mapstructure:"tunnelListenerSockets"`
	ExposedApps           []app.LocalExposedApp `mapstructure:"exposedApps"`
}
