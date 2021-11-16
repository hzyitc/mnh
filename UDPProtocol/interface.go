package UDPProtocol

import "net"

type Interface interface {
	ClosedChan() <-chan struct{}
	Close() error

	ServerAddr() net.Addr
	LocalHoleAddr() net.Addr
	NATedAddr() net.Addr
}
