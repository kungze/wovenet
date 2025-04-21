package message

import (
	"context"
	"fmt"
	"strings"

	"gihtub.com/kungze/wovenet/internal/logger"
)

type Payload struct {
	SiteName string      `json:"siteName"`
	ClientId string      `json:"clientId"`
	Kind     MessageKind `json:"kind"`
	Data     any         `json:"data"`
}

type MessageClient interface {
	// RegisterHandler register message handler
	RegisterHandler(kind MessageKind, cb Callback)
	// UnregisterHandler unregister message handler
	UnregisterHandler(kind MessageKind)
	// PublishMassage publish message to remote sites
	PublishMassage(ctx context.Context, msgKind MessageKind, data any) error
}

func NewMessageClient(ctx context.Context, config Config, siteName string) (MessageClient, error) {
	log := logger.GetDefault()
	log.Info("creating new message client", "protocol", config.Protocol)
	switch strings.ToLower(config.Protocol) {
	case MQTT:
		client, err := newMqttClient(ctx, *config.Mqtt, siteName, config.CryptoKey)
		if err != nil {
			log.Error("failed to create mqtt client", "error", err)
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}
}
