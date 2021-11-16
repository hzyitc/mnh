package UDPMode

import (
	"net"
)

const bufSize = 9000 - (20 + 8)

type Interface interface {
	Dial(serverAddr string) (net.Conn, error)
	ClosedChan() <-chan struct{}
	Close() error

	LocalHoleAddr() net.Addr
	ServiceAddr() net.Addr
}
