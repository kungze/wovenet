package message

type Config struct {
	Protocol  string     `mapstructure:"protocol"`
	CryptoKey string     `mapstructure:"cryptoKey"`
	Mqtt      mqttConfig `mapstructure:"mqtt"`
}
