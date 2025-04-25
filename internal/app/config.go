package app

import (
	"fmt"
	"net/netip"
	"slices"
	"strings"
)

type LocalExposedAppConfig struct {
	AppName   string `mapstructure:"appName"`
	AppSocket string `mapstructure:"appSocket"`
}

type RemoteAppConfig struct {
	SiteName    string `mapstructure:"siteName"`
	AppName     string `mapstructure:"appName"`
	LocalSocket string `mapstructure:"localSocket"`
}

var localExposedAppNames = []string{}

func CheckLocalExposedAppConfig(configs []*LocalExposedAppConfig) error {
	for _, cfg := range configs {
		if cfg.AppName == "" || cfg.AppSocket == "" {
			return fmt.Errorf("the appName and appsocket must be set together for local exposed app")
		}
		if len(cfg.AppName) > 255 {
			return fmt.Errorf("the appName: %s is too long, the max length is 255", cfg.AppName)
		}
		s := strings.SplitN(cfg.AppSocket, ":", 2)
		if len(s) != 2 {
			return fmt.Errorf("the appSocket: %s is invalid, the format must be <protocol>:<socket-addr>", cfg.AppSocket)
		}
		if !slices.Contains([]string{"tcp", "unix"}, s[0]) {
			return fmt.Errorf("the appSocket: %s is invalid, the protocol must be tcp or unix", cfg.AppSocket)
		}
		if strings.ToLower(s[0]) == "tcp" {
			_, err := netip.ParseAddrPort(s[1])
			if err != nil {
				return fmt.Errorf("the appSocket: %s is invalid", cfg.AppSocket)
			}
		}
		if slices.Contains(localExposedAppNames, cfg.AppName) {
			return fmt.Errorf("the appName: %s is duplicated", cfg.AppName)
		}
		localExposedAppNames = append(localExposedAppNames, cfg.AppName)
	}
	return nil
}

func CheckRemoteAddConfig(configs []*RemoteAppConfig) error {
	for _, cfg := range configs {
		if cfg.SiteName == "" || cfg.AppName == "" || cfg.LocalSocket == "" {
			return fmt.Errorf("the siteName, appName and localSocket must be set together for remote app")
		}
		if len(cfg.SiteName) > 255 {
			return fmt.Errorf("the siteName: %s is too long, the max length is 255", cfg.SiteName)
		}
		if len(cfg.AppName) > 255 {
			return fmt.Errorf("the appName: %s is too long, the max length is 255", cfg.AppName)
		}
		s := strings.SplitN(cfg.LocalSocket, ":", 2)
		if len(s) != 2 {
			return fmt.Errorf("the localSocket: %s is invalid, the format must be <protocol>:<socket-addr>", cfg.LocalSocket)
		}
		if !slices.Contains([]string{"tcp", "unix"}, s[0]) {
			return fmt.Errorf("the localSocket: %s is invalid, the protocol must be tcp or unix", cfg.LocalSocket)
		}
		if strings.ToLower(s[0]) == "tcp" {
			_, err := netip.ParseAddrPort(s[1])
			if err != nil {
				return fmt.Errorf("the localSocket: %s is invalid", cfg.LocalSocket)
			}
		}
	}
	return nil
}
