package site

import (
	"gihtub.com/kungze/wovenet/internal/app"
	"gihtub.com/kungze/wovenet/internal/tunnel"
)

type siteInfo struct {
	TunnelListenerSockets []tunnel.SocketInfo   `mapstructure:"tunnelListenerSockets"`
	ExposedApps           []app.LocalExposedApp `mapstructure:"exposedApps"`
}
