package tunnel

import (
	"context"
	"io"
	"net"
)

// Stream is a bidirectional stream, it's opened in a tunnel
type Stream io.ReadWriteCloser

// NewStreamCallback is called when a new stream is opened.
// the open operation is triggered by external client which
// try to connect to remote app or local exposed app
type NewStreamCallback func(stream Stream)

// RemoteSiteDisconnectedCallback is called when a remote site is disconnected.
// It is used to notify the local site that the remote site is no longer reachable.
type RemoteSiteDisconnectedCallback func(remoteSite string)

// RemoteSiteConnectedCallback is called when a remote site is connected.
// It is used to notify the local site that the remote site is reachable.
type RemoteSiteConnectedCallback func(ctx context.Context, remoteSite string)

// RequestNewRemoteSocketInfo is called whan a remote site's socket is disconnected
// and the remote socket's public address is dynamic. This function is used to
// request a new public address for the remote site.
type RequestNewRemoteSocketInfo func(remoteSite string, socketId string) error

// ConnectionErrorCallback is called when a connection error occurs.
// RequestNewRemoteSocketInfo will be called in this callback function
type ConnectionErrorCallback func(err error)

// DataChannelCreatedCallback is called when a new data channel is created.
// The tunnel between different sites is made up of one or more data channels.
// This callback function will add the data channel to the tunnel.
type DataChannelCreatedCallback func(remoteSite string, dc *dataChannel)

// DataChannelDestroyCallback is called when a data channel is destroyed.
// This callback function will remove the data channel from the tunnel.
type DataChannelDestroyCallback func(remoteSite string, channelId string)

type Listener interface {
	// Accept returns a new connections. It should be called in a loop.
	Accept(context.Context) (Connection, error)
	// Addr returns the local network addr that the server is listening on.
	Addr() net.Addr
}

type Dialer interface {
	// Dial request to establish a new tunnel connection with remote site
	Dial(context.Context) (Connection, error)
}

type Connection interface {
	// OpenStream opens a new bidirectional stream.
	OpenStream(context.Context) (Stream, error)

	// AcceptStream returns the next stream opened by the peer, blocking until one is available.
	AcceptStream(context.Context) (Stream, error)

	// LocalAddr returns the local network address, if known.
	LocalAddr() net.Addr

	// RemoteAddr returns the remote network address, if known.
	RemoteAddr() net.Addr

	Close()
}
