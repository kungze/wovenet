package app

import (
	"fmt"
)

type LocalExposedAppConfig struct {
	Mode         string   `json:"mode" mapstructure:"mode"`
	AppName      string   `json:"appName" mapstructure:"appName"`
	AppSocket    string   `json:"appSocket" mapstructure:"appSocket"`
	PortRange    []string `json:"portRange" mapstructure:"portRange"`
	AddressRange []string `json:"addressRange" mapstructure:"addressRange"`
}

type RemoteAppConfig struct {
	SiteName    string `json:"siteName" mapstructure:"siteName"`
	AppName     string `json:"appName" mapstructure:"appName"`
	LocalSocket string `json:"localSocket" mapstructure:"localSocket"`
	AppSocket   string `json:"appSocket" mapstructure:"appSocket"`
}

func CheckLocalExposedAppConfig(configs []*LocalExposedAppConfig) error {
	appNames := make(map[string]bool)

	for _, cfg := range configs {
		if len(cfg.AppName) == 0 {
			return fmt.Errorf("appName is required")
		}
		if len(cfg.AppName) > 255 {
			return fmt.Errorf("appName '%s' exceeds 255 characters", cfg.AppName)
		}
		if appNames[cfg.AppName] {
			return fmt.Errorf("duplicate appName found: '%s'", cfg.AppName)
		}
		appNames[cfg.AppName] = true

		if cfg.Mode == "" {
			cfg.Mode = SINGLE
		}

		switch cfg.Mode {
		case SINGLE:
			if cfg.AppSocket == "" {
				return fmt.Errorf("app '%s' has mode 'single' but missing appSocket", cfg.AppName)
			}
			err := isValidSocket(cfg.AppSocket)
			if err != nil {
				return fmt.Errorf("app '%s' has invalid appSocket: %s, %w", cfg.AppName, cfg.AppSocket, err)
			}
		case RANGE:
			if len(cfg.PortRange) == 0 {
				return fmt.Errorf("app '%s' has mode 'range' but missing portRange", cfg.AppName)
			}
			if err := validatePortRange(cfg.PortRange); err != nil {
				return fmt.Errorf("app '%s' has invalid portRange: %w", cfg.AppName, err)
			}
			if len(cfg.AddressRange) == 0 {
				return fmt.Errorf("app '%s' has mode 'range' but missing addressRange", cfg.AppName)
			}
			if err := validateAddressRange(cfg.AddressRange); err != nil {
				return fmt.Errorf("app '%s' has invalid addressRange: %w", cfg.AppName, err)
			}
		default:
			return fmt.Errorf("app '%s' has invalid mode: %s", cfg.AppName, cfg.Mode)
		}
	}
	return nil
}

func CheckRemoteAddConfig(configs []*RemoteAppConfig) error {
	for _, cfg := range configs {
		if cfg.SiteName == "" || cfg.AppName == "" || cfg.LocalSocket == "" {
			return fmt.Errorf("the siteName, appName and localSocket must be set together for remote app")
		}
		if len(cfg.SiteName) > 255 {
			return fmt.Errorf("the siteName: '%s' is too long, the max length is 255", cfg.SiteName)
		}
		if len(cfg.AppName) > 255 {
			return fmt.Errorf("the appName: '%s' is too long, the max length is 255", cfg.AppName)
		}
		if err := isValidSocket(cfg.LocalSocket); err != nil {
			return fmt.Errorf("the localSocket: '%s' of remote app: '%s' is invalid, %w", cfg.LocalSocket, cfg.AppName, err)
		}
		if cfg.AppSocket != "" {
			if err := isValidSocket(cfg.AppSocket); err != nil {
				return fmt.Errorf("the appSocket: '%s' of remote app: '%s' is invalid, %w", cfg.AppSocket, cfg.AppName, err)
			}
		}
	}
	return nil
}
