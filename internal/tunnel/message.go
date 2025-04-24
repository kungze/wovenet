package tunnel

type SocketInfo struct {
	Address              string            `mapstructure:"address"`
	Port                 uint16            `mapstructure:"port"`
	Protocol             TransportProtocol `mapstructure:"protocol"`
	Id                   string            `mapstructure:"id"`
	DynamicPublicAddress bool              `mapstructure:"dynamicPublicAddress"`
}

type SocketInfoRequest struct {
	Id string `mapstructure:"id"`
}
