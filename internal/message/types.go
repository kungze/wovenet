package message

type MessageKind string

type Callback func(*Payload) (any, MessageKind, error)
