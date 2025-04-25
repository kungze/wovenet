package crypto

import "fmt"

type Config struct {
	Key string `mapstructure:"key"`
}

func CheckConfig(config *Config) error {
	if config == nil {
		return fmt.Errorf("the crypto config is nil")
	}
	if config.Key == "" {
		return fmt.Errorf("the key must be set")
	}
	if len(config.Key) < 8 {
		return fmt.Errorf("the key is too short, the min length is 8")
	}
	if len(config.Key) > 255 {
		return fmt.Errorf("the key is too long, the max length is 255")
	}
	return nil
}
