package tunnel

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"time"

	"gihtub.com/kungze/wovenet/internal/logger"
	"github.com/quic-go/quic-go"
)

type quicConn struct {
	quic.Connection
}

// OpenStream opens a new bidirectional stream.
func (qc *quicConn) OpenStream(ctx context.Context) (Stream, error) {
	return qc.Connection.OpenStreamSync(ctx)
}

// AcceptStream returns the next stream opened by the peer, blocking until one is available.
func (qc *quicConn) AcceptStream(ctx context.Context) (Stream, error) {
	return qc.Connection.AcceptStream(ctx)
}

func (qc *quicConn) Close() {
}

type QuicDialer struct {
	socket SocketInfo
}

// Dial request to establish a new tunnel connection with remote site
func (qd *QuicDialer) Dial(ctx context.Context) (Connection, error) {
	log := logger.GetDefault()
	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"wovenet"},
	}
	addr := fmt.Sprintf("%s:%d", qd.socket.Address, qd.socket.Port)
	conn, err := quic.DialAddr(ctx, addr, tlsConf, &quic.Config{KeepAlivePeriod: 5 * time.Second})
	if err != nil {
		log.Error("failed to dial remote site", "remoteAddr", addr)
		return nil, err
	}

	return &quicConn{Connection: conn}, nil
}

func newQuicDialer(socket SocketInfo) *QuicDialer {
	return &QuicDialer{
		socket: socket,
	}
}

type QuicListener struct {
	*quic.Listener
	Config      SocketConfig
	Connections map[string]quic.Connection
}

func (qs *QuicListener) Accept(ctx context.Context) (Connection, error) {
	log := logger.GetDefault()
	conn, err := qs.Listener.Accept(ctx)
	if err != nil {
		log.Error("quic listener encountering error while accepting", "socket", qs.Addr().String(), "error", err)
		return nil, err
	}
	return &quicConn{Connection: conn}, nil
}

func (qs *QuicListener) GetSocketInfo() (*SocketInfo, error) {
	log := logger.GetDefault()
	var address string
	var port int
	switch qs.Config.PublicAddress {
	case STUN, HTTP:
		log.Error("unsuportted method for get public address", "method", qs.Config.PublicAddress)
		return nil, fmt.Errorf("not implement")
	default:
		address = qs.Config.PublicAddress
		port = qs.Config.PublicePort
	}
	return &SocketInfo{
		Address:  address,
		Port:     port,
		Protocol: qs.Config.TransportProtocol,
	}, nil
}

func (qs *QuicListener) Addr() net.Addr {
	return qs.Listener.Addr()
}

func newQuicListener(config SocketConfig) (*QuicListener, error) {
	log := logger.GetDefault()
	qListener := &QuicListener{
		Config: config,
	}
	switch config.Mode {
	case NatTraversal:
		log.Warn("net traversal mode have not implement yet")
		return nil, fmt.Errorf("QUIC does not support NAT traversal")
	case PortForwarding, DedicatedAddress:
		addr := fmt.Sprintf("%s:%d", config.ListenAddress, config.ListenPort)
		listener, err := quic.ListenAddr(addr, generateTLSConfig(), &quic.Config{})
		if err != nil {
			log.Error("failed to listen addr", "addr", addr, "error", err)
			return nil, err
		}
		qListener.Listener = listener
	default:
		log.Error("unsupported socket mode", "mode", config.Mode)
		return nil, fmt.Errorf("unsupported socket mode: %s", config.Mode)
	}

	return qListener, nil
}

// Setup a bare-bones TLS config for the server
func generateTLSConfig() *tls.Config {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"wovenet"},
	}
}
