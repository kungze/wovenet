package site

import (
	"gihtub.com/kungze/wovenet/internal/app"
	"gihtub.com/kungze/wovenet/internal/message"
	"gihtub.com/kungze/wovenet/internal/tunnel"
)

type Config struct {
	SiteName         string                      `mapstructure:"siteName"`
	MessageChannel   message.Config              `mapstructure:"messageChannel"`
	Tunnel           tunnel.Config               `mapstructure:"tunnel"`
	Stun             []string                    `mapstructure:"stun"`
	LocalExposedApps []app.LocalExposedAppConfig `mapstructure:"localExposedApps"`
	RemoteApps       []app.RemoteAppConfig       `mapstructure:"remoteApps"`
}
