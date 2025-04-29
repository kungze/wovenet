package tunnel

import (
	"context"
	"sync"

	"github.com/google/uuid"
	"github.com/kungze/wovenet/internal/logger"
)

// dataChannel wraps different connection types(such as QUIC connection)
type dataChannel struct {
	Connection
	id                string
	active            bool
	mux               sync.Mutex
	remoteSite        string
	ctx               context.Context
	cancel            context.CancelFunc
	streamCallback    NewStreamCallback
	onConnectionError ConnectionErrorCallback
	destroyCallback   DataChannelDestroyCallback
}

func (dc *dataChannel) GetId() string {
	return dc.id
}

func (dc *dataChannel) IsActive() bool {
	dc.mux.Lock()
	defer dc.mux.Unlock()
	return dc.active
}

func (dc *dataChannel) Start() {
	dc.mux.Lock()
	defer dc.mux.Unlock()
	if dc.active {
		logger.GetDefault().Error("the data channel is already active", "remoteSite", dc.remoteSite)
		return
	}
	dc.acceptStreamLoop()
}

func (dc *dataChannel) Destroy() {
	dc.cancel()
	go dc.destroyCallback(dc.remoteSite, dc.id)
}

// acceptStreamLoop is used to accept stream from remote site.
// the stream used to transfer data between remote site's external
// client and local exposed app.
func (dc *dataChannel) acceptStreamLoop() {
	dc.active = true
	log := logger.GetDefault()
	go func() {
		log.Info("the data channel start to accept stream", "dataChannelId", dc.id, "remoteSite", dc.remoteSite)
		defer func() {
			dc.mux.Lock()
			dc.active = false
			dc.mux.Unlock()
			dc.Close()
		}()
		for {
			select {
			case <-dc.ctx.Done():
				return
			default:
				stream, err := dc.AcceptStream(dc.ctx)
				if err != nil {
					log.Error("failed to accept stream", "remoteSite", dc.remoteSite, "error", err)
					if dc.onConnectionError != nil {
						go dc.onConnectionError(err)
					}
					dc.Destroy()
					return
				}
				log.Info("accept a new stream", "remoteSite", dc.remoteSite)
				go dc.streamCallback(stream)
			}
		}
	}()
}

func newDataChannel(ctx context.Context, conn Connection, remoteSite string,
	streamCallback NewStreamCallback,
	onConnectionError ConnectionErrorCallback,
	destroyCallback DataChannelDestroyCallback) *dataChannel {
	dc := &dataChannel{
		Connection:        conn,
		id:                uuid.NewString(),
		active:            false,
		remoteSite:        remoteSite,
		streamCallback:    streamCallback,
		onConnectionError: onConnectionError,
		destroyCallback:   destroyCallback,
		mux:               sync.Mutex{},
	}
	dc.ctx, dc.cancel = context.WithCancel(ctx)
	return dc
}
