package message

const (
	MQTT string = "mqtt"
)

const (
	ExchangeInfoRequest   MessageKind = "exchange-message-request"
	ExchangeInfoResponse  MessageKind = "exchange-message-response"
	NewSocketInfoRequest  MessageKind = "new-socket-info-request"
	NewSocketInfoResponse MessageKind = "new-socket-info-response"
)
