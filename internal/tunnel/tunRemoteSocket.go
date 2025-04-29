package tunnel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gihtub.com/kungze/wovenet/internal/logger"
)

type tunRemoteSocket struct {
	SocketInfo
	mux                        sync.Mutex
	dataChannel                *dataChannel
	remoteSite                 string
	localSite                  string
	ctx                        context.Context
	cancel                     context.CancelFunc
	requestNewSocketInfo       RequestNewRemoteSocketInfo
	streamCallback             NewStreamCallback
	dataChannelCreatedCallback DataChannelCreatedCallback
	dataChannelDestroyCallback DataChannelDestroyCallback
}

func (s *tunRemoteSocket) Destroy() {
	if s.dataChannel != nil {
		s.dataChannel.Destroy()
	}
	s.cancel()
}

func (s *tunRemoteSocket) OpenDataChannel(ctx context.Context) error {
	s.mux.Lock()
	defer s.mux.Unlock()
	log := logger.GetDefault()
	if s.dataChannel != nil && s.dataChannel.IsActive() {
		log.Info("there is already an active data channel", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo)
		go s.dataChannelCreatedCallback(s.remoteSite, s.dataChannel)
		return nil
	}
	var dialer Dialer
	switch s.Protocol {
	case QUIC:
		dialer = newQuicDialer(s.SocketInfo)
	case SCTP:
		return fmt.Errorf("unsupported protocol: %s", SCTP)
	default:
		return fmt.Errorf("unsupported protocol: %s", s.Protocol)
	}

	conn, err := dialer.Dial(ctx)
	if err != nil {
		log.Error("Failed to dial remote site", "remoteSite", s.remoteSite, "remoteAddr", fmt.Sprintf("%s:%d", s.Address, s.Port))
		return err
	}
	log.Info("connect to remote site", "remoteSite", s.remoteSite, "remoteAddr", fmt.Sprintf("%s:%d", s.Address, s.Port))
	// open a control stream, we will tell remote site the local site name by this control stream
	stream, err := conn.OpenStream(ctx)
	if err != nil {
		log.Error("failed to open control stream", "remoteSite", s.remoteSite, "error", err)
		conn.Close()
		return err
	}
	data := []byte(s.localSite)
	len := byte(len(data))
	n, err := stream.Write(append([]byte{len}, data...))
	if err != nil {
		log.Error("failed to write date to control stream", "remoteSite", s.remoteSite, "error", err)
		_ = stream.Close()
		conn.Close()
		return err
	}
	if n != int(len)+1 {
		stream.Close() //nolint:errcheck
		conn.Close()
		log.Error("the length of data write to control stream is valid", "remoteSite", s.remoteSite)
		return fmt.Errorf("write data length is not valid")
	}
	s.dataChannel = newDataChannel(ctx, conn, s.remoteSite, s.streamCallback, s.onConnectionError, s.dataChannelDestroyCallback)
	s.dataChannel.Start()
	go s.dataChannelCreatedCallback(s.remoteSite, s.dataChannel)
	return nil
}

// onConnectionError is called when the a data channel encounter an error
// it will try to reconnect to the remote socket.
func (s *tunRemoteSocket) onConnectionError(err error) {
	log := logger.GetDefault()
	if s.DynamicPublicAddress {
		// if the socket is dynamic public address, we will try to request a new socket info
		// from remote site, so that the local site can connect to the remote site
		// by this new socket
		duration := 1 * time.Second
		for {
			select {
			case <-s.ctx.Done():
				log.Info("the context is done, stop requesting new socket info", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo)
				return
			case <-time.NewTicker(duration).C:
				if s.dataChannel != nil && s.dataChannel.IsActive() {
					log.Info("the remote socket is connected again", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo)
					return
				}
				log.Info("the remote socket is not connected, try to request new socket info again", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo)
				if err := s.requestNewSocketInfo(s.remoteSite, s.Id); err != nil {
					log.Error("failed to request new remote socket info", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo, "error", err)
				}
				if duration < 5*time.Minute {
					duration += 5 * time.Second
				}
			}
		}
	} else {
		// if the socket is not dynamic public address, we will try to dial the remote site directly
		duration := 1 * time.Second
		for {
			select {
			case <-s.ctx.Done():
				log.Info("the context is done, stop dialing remote site", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo)
				return
			case <-time.NewTicker(duration).C:
				if s.dataChannel != nil && s.dataChannel.IsActive() {
					log.Info("the remote socket is connected again", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo)
					return
				}
				log.Info("the remote socket is not connected, try to dial again", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo)
				if err := s.OpenDataChannel(s.ctx); err != nil {
					log.Error("failed to dial remote site", "remoteSite", s.remoteSite, "socketInfo", s.SocketInfo, "error", err)
				}
				if duration < 5*time.Minute {
					duration += 5 * time.Second
				}
			}
		}
	}
}

func newTunRemoteSocket(
	ctx context.Context, info SocketInfo, localSite, remoteSite string,
	requestNewSocketInfo RequestNewRemoteSocketInfo, streamCallback NewStreamCallback,
	dataChannelCreatedCallback DataChannelCreatedCallback,
	dataChannelDestroyCallback DataChannelDestroyCallback) *tunRemoteSocket {
	ctx, cancel := context.WithCancel(ctx)
	return &tunRemoteSocket{
		SocketInfo:                 info,
		mux:                        sync.Mutex{},
		localSite:                  localSite,
		remoteSite:                 remoteSite,
		ctx:                        ctx,
		cancel:                     cancel,
		requestNewSocketInfo:       requestNewSocketInfo,
		streamCallback:             streamCallback,
		dataChannelCreatedCallback: dataChannelCreatedCallback,
		dataChannelDestroyCallback: dataChannelDestroyCallback,
	}
}
