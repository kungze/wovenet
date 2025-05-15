package tunnel

type SocketInfo struct {
	Address              string            `json:"address" mapstructure:"address"`
	Port                 uint16            `json:"port" mapstructure:"port"`
	Protocol             TransportProtocol `json:"protocol" mapstructure:"protocol"`
	Id                   string            `json:"id" mapstructure:"id"`
	DynamicPublicAddress bool              `json:"dynamicPublicAddress" mapstructure:"dynamicPublicAddress"`
}

type SocketInfoRequest struct {
	Id string `mapstructure:"id"`
}
