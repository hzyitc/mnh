package TCPMode

import (
	"net"
	"strconv"

	"github.com/libp2p/go-reuseport"
)

type Listener interface {
	Interface
	net.Listener
}

type listener struct {
	port int

	closingChan chan struct{}
	closedChan  chan struct{}

	net.Listener
	reuse Interface
}

func NewListener(rfc string, port int) (Listener, error) {
	local := "0.0.0.0:" + strconv.Itoa(port)
	server, err := reuseport.Listen("tcp", local)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveTCPAddr("tcp", server.Addr().String())
	if err != nil {
		server.Close()
		return nil, err
	}
	port = addr.Port

	reuse, err := NewReuse(rfc, port)
	if err != nil {
		server.Close()
		return nil, err
	}

	s := &listener{
		port,

		make(chan struct{}),
		make(chan struct{}),

		server,
		reuse,
	}

	return s, nil
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
	err2 := s.Listener.Close()

	close(s.closedChan)
	if err != nil {
		return err
	} else {
		return err2
	}
}

func (s *listener) LocalHoleAddr() net.Addr {
	return s.reuse.LocalHoleAddr()
}

func (s *listener) ServiceAddr() net.Addr {
	return s.Listener.Addr()
}
