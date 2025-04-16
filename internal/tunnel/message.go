package tunnel

type SocketInfo struct {
	Address  string            `mapstructure:"address"`
	Port     int               `mapstructure:"port"`
	Protocol TransportProtocol `mapstructure:"protocol"`
}
