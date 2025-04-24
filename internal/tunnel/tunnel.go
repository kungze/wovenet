package tunnel

import (
	"context"
	"io"
	"math/rand"
	"sync"

	"gihtub.com/kungze/wovenet/internal/logger"
	"github.com/google/uuid"
)

type tunnelBrokenCallback func()

type tunnel struct {
	slaves         map[string]Connection
	streamCallback NewStreamCallback
	brokenCallback tunnelBrokenCallback
	mux            sync.RWMutex
	remoteSite     string
}

func (t *tunnel) OpenStream(ctx context.Context) (io.ReadWriteCloser, error) {
	t.mux.RLock()
	defer t.mux.RUnlock()
	conns := []Connection{}
	for _, conn := range t.slaves {
		conns = append(conns, conn)
	}
	if len(conns) > 1 {
		return conns[rand.Intn(len(t.slaves)-1)].OpenStream(ctx)
	} else {
		return conns[0].OpenStream(ctx)
	}
}

func (t *tunnel) delSlave(key string) {
	log := logger.GetDefault()
	t.mux.Lock()
	defer t.mux.Unlock()
	delete(t.slaves, key)
	if len(t.slaves) == 0 {
		log.Warn("all slave data channels are disconnected", "remoteSite", t.remoteSite)
		t.brokenCallback()
	}
}

func (t *tunnel) addSlaveConn(ctx context.Context, conn Connection) {
	t.mux.Lock()
	defer t.mux.Unlock()
	log := logger.GetDefault()
	key := uuid.NewString()
	t.slaves[key] = conn
	go func() {
		defer conn.Close()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				stream, err := conn.AcceptStream(ctx)
				if err != nil {
					log.Warn("encountering error while accepting stream", "error", err)
					t.delSlave(key)
					return
				} else {
					log.Info("accept a new stream", "remoteAddr", conn.RemoteAddr().String())
					go t.streamCallback(stream)
				}
			}
		}
	}()
}

func newTunnel(remoteSite string, streamCallback NewStreamCallback, brokenCallback tunnelBrokenCallback) *tunnel {
	return &tunnel{
		remoteSite:     remoteSite,
		slaves:         map[string]Connection{},
		streamCallback: streamCallback,
		brokenCallback: brokenCallback,
		mux:            sync.RWMutex{},
	}
}
