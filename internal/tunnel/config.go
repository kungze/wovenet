package tunnel

type SocketConfig struct {
	Mode              SocketMode        `mapstructure:"mode"`
	TransportProtocol TransportProtocol `mapstructure:"transportProtocol"`
	PublicAddress     string            `mapstructure:"publicAddress"`
	PublicePort       int               `mapstructure:"publicPort"`
	ListenAddress     string            `mapstructure:"listenAddress"`
	ListenPort        int               `mapstructure:"listenPort"`
}

type Config struct {
	LocalSockets []SocketConfig `mapstructure:"localSockets"`
}
