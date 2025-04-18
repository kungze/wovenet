package message

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
)

type Config struct {
	Protocol  string     `mapstructure:"protocol"`
	CryptoKey string     `mapstructure:"cryptoKey"`
	Mqtt      mqttConfig `mapstructure:"mqtt"`
}

func CheckAndSetDefaultConfig(config Config) (*Config, error) {
	if !slices.Contains([]string{MQTT}, strings.ToLower(config.Protocol)) {
		return nil, fmt.Errorf("unsupported message protocol: %s", config.Protocol)
	}
	if len(config.CryptoKey) != 16 {
		return nil, fmt.Errorf("the expected length of message cryptoKey is 16, but got lenght: %d", len(config.CryptoKey))
	}
	if config.Protocol == strings.ToLower(MQTT) {
		if config.Mqtt.Topic == "" {
			return nil, fmt.Errorf("the mqtt topic must be set")
		}
		if config.Mqtt.BrokerServer == "" {
			config.Mqtt.BrokerServer = "mqtt://mqtt.eclipseprojects.io:1883"
		}
		if !strings.HasPrefix(config.Mqtt.BrokerServer, "mqtt://") {
			return nil, fmt.Errorf("the mqtt brokerServer must start with 'mqtt://'")
		}
		_, err := url.Parse(config.Mqtt.BrokerServer)
		if err != nil {
			return nil, fmt.Errorf("the brokerServer: %s is not a valid url", config.Mqtt.BrokerServer)
		}
	}
	return &config, nil
}
