package site

import (
	"fmt"

	"gihtub.com/kungze/wovenet/internal/app"
	"gihtub.com/kungze/wovenet/internal/crypto"
	"gihtub.com/kungze/wovenet/internal/message"
	"gihtub.com/kungze/wovenet/internal/tunnel"
)

type Config struct {
	SiteName         string                       `mapstructure:"siteName"`
	Crypto           *crypto.Config               `mapstructure:"crypto"`
	MessageChannel   *message.Config              `mapstructure:"messageChannel"`
	Tunnel           *tunnel.Config               `mapstructure:"tunnel"`
	Stun             []string                     `mapstructure:"stun"`
	LocalExposedApps []*app.LocalExposedAppConfig `mapstructure:"localExposedApps"`
	RemoteApps       []*app.RemoteAppConfig       `mapstructure:"remoteApps"`
}

func CheckAndSetDefaultConfig(config *Config) error {
	if config.SiteName == "" {
		return fmt.Errorf("the siteName must be set")
	}
	if len(config.SiteName) > 255 {
		return fmt.Errorf("the siteName is too long, the lenght must less or equal 255")
	}

	if err := crypto.CheckConfig(config.Crypto); err != nil {
		return err
	}

	if err := message.CheckAndSetDefaultConfig(config.MessageChannel); err != nil {
		return err
	}

	if err := tunnel.CheckAndSetDefaultConfig(config.Tunnel); err != nil {
		return err
	}

	if err := app.CheckLocalExposedAppConfig(config.LocalExposedApps); err != nil {
		return err
	}
	if err := app.CheckRemoteAddConfig(config.RemoteApps); err != nil {
		return err
	}
	return nil
}
