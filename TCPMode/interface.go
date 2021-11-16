package TCPMode

import (
	"net"
)

type Interface interface {
	Dial(serverAddr string) (net.Conn, error)
	ClosedChan() <-chan struct{}
	Close() error

	LocalHoleAddr() net.Addr
	ServiceAddr() net.Addr
}
