package tunnel

type SocketInfo struct {
	Address  string            `mapstructure:"address"`
	Port     uint16            `mapstructure:"port"`
	Protocol TransportProtocol `mapstructure:"protocol"`
}
