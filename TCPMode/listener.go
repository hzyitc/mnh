package TCPMode

import (
	"net"
	"strconv"

	"github.com/libp2p/go-reuseport"

	"github.com/hzyitc/mnh/routerPortForward"
)

type listener struct {
	port int

	closingChan chan struct{}
	closedChan  chan struct{}

	server net.Listener
	reuse  Interface
}

func NewListener(rpfc routerPortForward.Config, port int) (Interface, net.Listener, error) {
	local := "0.0.0.0:" + strconv.Itoa(port)
	server, err := reuseport.Listen("tcp", local)
	if err != nil {
		return nil, nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", server.Addr().String())
	if err != nil {
		server.Close()
		return nil, nil, err
	}
	port = addr.Port

	reuse, err := NewReuse(rpfc, port)
	if err != nil {
		server.Close()
		return nil, nil, err
	}

	s := &listener{
		port,

		make(chan struct{}),
		make(chan struct{}),

		server,
		reuse,
	}

	return s, server, nil
}

func (s *listener) Dial(addr string) (net.Conn, error) {
	return s.reuse.Dial(addr)
}

func (s *listener) ClosedChan() <-chan struct{} {
	return s.closedChan
}

func (s *listener) Close() error {
	select {
	case <-s.closingChan:
		return nil
	default:
		break
	}
	close(s.closingChan)

	err := s.reuse.Close()
	err2 := s.server.Close()

	close(s.closedChan)
	if err != nil {
		return err
	} else {
		return err2
	}
}

func (s *listener) LocalServiceAddr() net.Addr {
	return &net.TCPAddr{
		Port: s.port,
	}
}
