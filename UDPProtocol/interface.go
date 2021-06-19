package UDPProtocol

import "net"

type Interface interface {
	ClosedChan() <-chan struct{}
	Close() error

	RemoteServerAddr() net.Addr
	LocalHoleAddr() net.Addr
	RemoteHoleAddr() net.Addr
}
