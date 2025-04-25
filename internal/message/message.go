package message

import (
	"context"
	"fmt"
	"strings"

	"gihtub.com/kungze/wovenet/internal/crypto"
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
	// BroadcastMessage broadcast message to all sites
	BroadcastMessage(ctx context.Context, msgKind MessageKind, data any) error
	// UnicastMessage send message to a specific site
	UnicastMessage(ctx context.Context, siteName string, msgKind MessageKind, data any) error
}

func NewMessageClient(ctx context.Context, config Config, cryptoConfig crypto.Config, siteName string) (MessageClient, error) {
	log := logger.GetDefault()
	log.Info("creating new message client", "protocol", config.Protocol)
	switch strings.ToLower(config.Protocol) {
	case MQTT:
		client, err := newMqttClient(ctx, *config.Mqtt, cryptoConfig, siteName)
		if err != nil {
			log.Error("failed to create mqtt client", "error", err)
			return nil, err
		}
		return client, nil
	default:
		return nil, fmt.Errorf("unsupported protocol: %s", config.Protocol)
	}
}
