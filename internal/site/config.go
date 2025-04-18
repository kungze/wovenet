package site

import (
	"fmt"

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

func CheckAndSetDefaultConfig(config Config) (*Config, error) {
	if config.SiteName == "" {
		return nil, fmt.Errorf("the siteName must be set")
	}
	if len(config.SiteName) > 255 {
		return nil, fmt.Errorf("the siteName is too long, the lenght must less or equal 255")
	}
	msgCfg, err := message.CheckAndSetDefaultConfig(config.MessageChannel)
	if err != nil {
		return nil, err
	}
	config.MessageChannel = *msgCfg
	tunCfg, err := tunnel.CheckAndSetDefaultConfig(config.Tunnel)
	if err != nil {
		return nil, err
	}
	config.Tunnel = *tunCfg
	if err := app.CheckLocalExposedAppConfig(config.LocalExposedApps); err != nil {
		return nil, err
	}
	if err := app.CheckRemoteAddConfig(config.RemoteApps); err != nil {
		return nil, err
	}
	return &config, nil
}
