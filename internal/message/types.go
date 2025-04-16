package message

type MessageKind string

const (
	ExchangeInfoRequest  MessageKind = "exchange-message-request"
	ExchangeInfoResponse MessageKind = "exchange-message-response"
)

type Callback func(*Payload) (any, MessageKind, error)
