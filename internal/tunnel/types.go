package tunnel

import (
	"context"
	"io"
	"net"
)

type Stream io.ReadWriteCloser

type RemoteSiteGoneCallback func(remoteSite string)
type RemoteSiteConnectedCallback func(ctx context.Context, remoteSite string) error

type NewStreamCallback func(stream Stream)

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
