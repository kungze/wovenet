package app

import (
	"fmt"
	"net/netip"
	"slices"
)

type LocalExposedAppConfig struct {
	Id     string `mapstructure:"id"`
	Socket string `mapstructure:"socket"`
}

type RemoteAppConfig struct {
	RemoteAppId string `mapstructure:"remoteAppId"`
	LocalSocket string `mapstructure:"localSocket"`
	SiteName    string `mapstructure:"siteName"`
}

var localExposedAppIds = []string{}

func CheckLocalExposedAppConfig(configs []*LocalExposedAppConfig) error {
	for _, cfg := range configs {
		if cfg.Id == "" || cfg.Socket == "" {
			return fmt.Errorf("the id and socket must set together for localExposedApp")
		}
		_, err := netip.ParseAddrPort(cfg.Socket)
		if err != nil {
			return fmt.Errorf("the socket: %s is invalid", cfg.Socket)
		}
		if slices.Contains(localExposedAppIds, cfg.Id) {
			return fmt.Errorf("the local exposed app id: %s is not unique", cfg.Id)
		}
		localExposedAppIds = append(localExposedAppIds, cfg.Id)
	}
	return nil
}

func CheckRemoteAddConfig(configs []*RemoteAppConfig) error {
	for _, cfg := range configs {
		if cfg.RemoteAppId == "" || cfg.LocalSocket == "" || cfg.SiteName == "" {
			return fmt.Errorf("the siteName and remoteAppId and localSocket must be set together for remoteApp")
		}
		_, err := netip.ParseAddrPort(cfg.LocalSocket)
		if err != nil {
			return fmt.Errorf("the localSocket: %s is invalid", cfg.LocalSocket)
		}

	}
	return nil
}
