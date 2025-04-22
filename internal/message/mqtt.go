package message

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"gihtub.com/kungze/wovenet/internal/logger"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
)

const (
	specificSiteTopic = "%s/%s/responses"
)

type mqttConfig struct {
	BrokerServer string `mapstructure:"brokerServer"`
	Topic        string `mapstructure:"topic"`
}

type mqttClient struct {
	mqttClient          *autopaho.ConnectionManager
	handlers            map[MessageKind]Callback
	clientId            string
	siteName            string
	topic               string
	cryptoKey           string
	siteNameClientIdMap sync.Map
}

// RegisterHandler register message handler
func (mc *mqttClient) RegisterHandler(kind MessageKind, cb Callback) {
	mc.handlers[kind] = cb
}

// UnregisterHandler unregister message handler
func (mc *mqttClient) UnregisterHandler(kind MessageKind) {
	delete(mc.handlers, kind)
}

// UnicastMessage send message to a specific site
func (mc *mqttClient) UnicastMessage(ctx context.Context, siteName string, msgKind MessageKind, data any) error {
	log := logger.GetDefault()
	clientId, ok := mc.siteNameClientIdMap.Load(siteName)
	if !ok {
		log.Error("can not found site client id", "siteName", siteName)
		return fmt.Errorf("site %s not found", siteName)
	}
	err := mc.publishMassage(ctx, fmt.Sprintf(specificSiteTopic, mc.topic, clientId), msgKind, data)
	if err != nil {
		log.Error("failed to publish message", "error", err)
		return err
	}
	return nil
}

// BroadcastMessage broadcast message to all sites
func (mc *mqttClient) BroadcastMessage(ctx context.Context, msgKind MessageKind, data any) error {
	log := logger.GetDefault()
	err := mc.publishMassage(ctx, mc.topic, msgKind, data)
	if err != nil {
		log.Error("failed to publish message", "error", err)
		return err
	}
	return nil
}

func (mc *mqttClient) publishMassage(ctx context.Context, topic string, msgKind MessageKind, data any) error {
	log := logger.GetDefault()
	payload := &Payload{
		SiteName: mc.siteName,
		ClientId: mc.clientId,
		Kind:     msgKind,
		Data:     data,
	}
	mData, err := json.Marshal(payload)
	if err != nil {
		log.Error("failed to marshal message payload", "error", err)
		return err
	}
	_, err = mc.mqttClient.Publish(ctx, &paho.Publish{
		QoS:     2,
		Topic:   topic,
		Payload: []byte(encrypt(mData, mc.cryptoKey)),
	})
	if err != nil {
		log.Error("failed to publish message", "error", err)
		return err
	}
	return nil
}

func (mc *mqttClient) onPublishReceived(r paho.PublishReceived) (bool, error) {
	log := logger.GetDefault()
	if r.AlreadyHandled {
		return true, nil
	}

	payload := &Payload{}
	if err := json.Unmarshal(decrypt(string(r.Packet.Payload), mc.cryptoKey), payload); err != nil {
		log.Error("failed to umarshal message", "error", err)
		return false, err
	}
	if payload.ClientId == mc.clientId {
		log.Warn("received message from the site self, ignore it")
		return false, nil
	}
	handler, ok := mc.handlers[payload.Kind]
	if !ok {
		log.Warn("can not found message handler", "messageKind", payload.Kind)
		return false, nil
	}
	resp, kind, err := handler(payload)
	if err != nil {
		log.Error("can not handle message payload", "messageKind", payload.Kind, "error", err)
		return false, err
	}
	// If resp is not nil, means that we need to responed to remote site
	if resp != nil {
		err := mc.publishMassage(context.Background(), fmt.Sprintf(specificSiteTopic, mc.topic, payload.ClientId), kind, resp)
		if err != nil {
			return false, err
		}
	}
	mc.siteNameClientIdMap.Store(payload.SiteName, payload.ClientId)

	return true, nil
}

func (mc *mqttClient) onError(err error) {
	log := logger.GetDefault()
	log.Error("message client encounter error", "error", err)
}

func newMqttClient(ctx context.Context, mqttConfig mqttConfig, siteName string, cryptoKey string) (*mqttClient, error) {
	log := logger.GetDefault()
	mClient := &mqttClient{
		siteName:            siteName,
		clientId:            uuid.NewString(),
		handlers:            make(map[MessageKind]Callback),
		topic:               mqttConfig.Topic,
		cryptoKey:           cryptoKey,
		siteNameClientIdMap: sync.Map{},
	}
	if mqttConfig.BrokerServer == "" {
		log.Error("broker server is empty")
		return nil, fmt.Errorf("broker server is empty")
	}
	u, err := url.Parse(mqttConfig.BrokerServer)
	if err != nil {
		log.Error("failed to parset brroker server url", "error", err)
		return nil, err
	}
	subscribes := []paho.SubscribeOptions{
		{Topic: mqttConfig.Topic, QoS: 2},
		{Topic: fmt.Sprintf(specificSiteTopic, mqttConfig.Topic, mClient.clientId), QoS: 2},
	}
	clientConfig := autopaho.ClientConfig{
		ServerUrls:       []*url.URL{u},
		ReconnectBackoff: autopaho.NewConstantBackoff(2 * time.Second),
		ConnectTimeout:   5 * time.Second,
		OnConnectionUp: func(cm *autopaho.ConnectionManager, c *paho.Connack) {
			log.Info("connect to MQTT broker successful")
			ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
			defer cancel()
			if _, err := cm.Subscribe(ctx, &paho.Subscribe{
				Subscriptions: subscribes,
			}); err != nil {
				log.Error("failed to subscribe mqtt message", "error", err)
				mClient.onError(err)
				return
			}
			log.Info("subscribe mqtt message successful")
		},
		OnConnectError: func(err error) {
			log.Error("error whilst attempting connection", "error", err)
			mClient.onError(err)
		},

		ClientConfig: paho.ClientConfig{
			ClientID:          mClient.clientId,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){mClient.onPublishReceived},
			OnClientError: func(err error) {
				log.Error("mqtt client error", "error", err)
				mClient.onError(err)
			},
			OnServerDisconnect: func(d *paho.Disconnect) {
				err := fmt.Errorf("server disconnected, code: %d, reason: %s", d.ReasonCode, d.Properties.ReasonString)
				log.Error(err.Error())
				mClient.onError(err)
			},
		},
	}

	cm, err := autopaho.NewConnection(ctx, clientConfig)
	if err != nil {
		return nil, err
	}
	// Wait for the connection to come up
	if err = cm.AwaitConnection(ctx); err != nil {
		return nil, err
	}
	mClient.mqttClient = cm
	return mClient, nil
}
