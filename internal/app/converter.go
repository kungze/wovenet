package app

import (
	"errors"
	"io"
	"net"
	"sync"

	"github.com/kungze/wovenet/internal/logger"
	"github.com/kungze/wovenet/internal/tunnel"
)

var errInvalidWrite = errors.New("invalid write result")

var buffPool *sync.Pool
var once sync.Once

func getBuffPool() *sync.Pool {
	once.Do(func() {
		buffPool = &sync.Pool{
			New: func() any {
				size := 32 * 1024
				buf := make([]byte, size)
				return &buf
			},
		}
	})
	return buffPool
}

type converter struct {
	conn       net.Conn
	stream     tunnel.Stream
	appType    AppType
	appName    string
	remoteSite string
}

func (c *converter) copy(dst io.Writer, src io.Reader) (err error) {
	bufPool := getBuffPool()
	bufPtr := bufPool.Get().(*[]byte)
	buf := *bufPtr
	defer bufPool.Put(bufPtr)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errInvalidWrite
				}
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return
}

func (c *converter) Start() {
	log := logger.GetDefault()
	wg := sync.WaitGroup{}
	wg.Add(2)
	go func() {
		defer func() {
			_ = c.conn.Close()
			_ = c.stream.Close()
			wg.Done()
		}()
		err := c.copy(c.conn, c.stream)
		if err != nil {
			log.Warn("failed to copy data from tunnel stream to tcp/unix connection", "appType", c.appType, "remoteSite", c.remoteSite, "appName", c.appName)
		}
	}()
	go func() {
		defer func() {
			_ = c.conn.Close()
			_ = c.stream.Close()
			wg.Done()
		}()
		err := c.copy(c.stream, c.conn)
		if err != nil {
			log.Warn("failed to copy data from tcp/unix connection to tunnel stream", "appType", c.appType, "remoteSite", c.remoteSite, "appName", c.appName)
		}
	}()
	wg.Wait()
	log.Info("converter exit", "appType", c.appType, "remoteSite", c.remoteSite, "appName", c.appName)
}
