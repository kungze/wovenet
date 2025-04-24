package tunnel

import (
	"context"
	"fmt"
	"io"
	"math/rand"
	"sync"

	"gihtub.com/kungze/wovenet/internal/logger"
)

type tunnelBrokenCallback func()

type tunnel struct {
	slaves         map[string]*dataChannel
	brokenCallback tunnelBrokenCallback
	mux            sync.RWMutex
	remoteSite     string
}

func (t *tunnel) OpenStream(ctx context.Context) (io.ReadWriteCloser, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	channnels := []*dataChannel{}
	for _, channel := range t.slaves {
		if channel.IsActive() {
			channnels = append(channnels, channel)
		}
	}
	if len(channnels) == 0 {
		return nil, fmt.Errorf("no available data channel")
	}
	if len(channnels) > 1 {
		return channnels[rand.Intn(len(t.slaves)-1)].OpenStream(ctx)
	} else {
		return channnels[0].OpenStream(ctx)
	}
}

func (t *tunnel) DeleteSlaveDataChannel(channelId string) {
	log := logger.GetDefault()
	t.mux.Lock()
	defer t.mux.Unlock()
	delete(t.slaves, channelId)
	if len(t.slaves) == 0 {
		log.Warn("all slave data channels are disconnected", "remoteSite", t.remoteSite)
		go t.brokenCallback()
	}
}

func (t *tunnel) AddSlaveDataChannel(dc *dataChannel) {
	t.mux.Lock()
	defer t.mux.Unlock()

	t.slaves[dc.GetId()] = dc

}

func newTunnel(remoteSite string, brokenCallback tunnelBrokenCallback) *tunnel {
	return &tunnel{
		remoteSite:     remoteSite,
		slaves:         make(map[string]*dataChannel),
		brokenCallback: brokenCallback,
		mux:            sync.RWMutex{},
	}
}
