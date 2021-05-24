package TCPMode

import (
	"errors"
	"net"
	"strconv"

	"github.com/libp2p/go-reuseport"

	"github.com/hzyitc/mnh/routerPortForward"
)

type reuse struct {
	port int
	rpf  routerPortForward.Interface

	closingChan chan struct{}
	closedChan  chan struct{}

	conn *reuseConn
}

type reuseConn struct {
	net.Conn

	reuse *reuse
}

func NewReuse(rpfc routerPortForward.Config, port int) (Interface, error) {
	rpf, err := routerPortForward.New(rpfc, port)
	if err != nil {
		return nil, err
	}

	return &reuse{
		port,
		rpf,

		make(chan struct{}),
		make(chan struct{}),

		nil,
	}, nil
}

func (s *reuse) Dial(addr string) (net.Conn, error) {
	if s.conn != nil {
		return nil, errors.New("double dial")
	}

	select {
	case <-s.closingChan:
		return nil, net.ErrClosed
	default:
		local := "0.0.0.0:" + strconv.Itoa(s.port)
		conn, err := reuseport.Dial("tcp", local, addr)
		if err != nil {
			return nil, err
		}

		c := &reuseConn{
			conn,
			s,
		}
		s.conn = c

		return c, nil
	}
}

func (s *reuse) ClosedChan() <-chan struct{} {
	return s.closedChan
}

func (s *reuse) Close() error {
	select {
	case <-s.closingChan:
		return nil
	default:
		break
	}
	close(s.closingChan)

	var err error
	if s.conn != nil {
		err = s.conn.Close()
	}
	s.rpf.Close()

	close(s.closedChan)
	return err
}

func (s *reuse) LocalServiceAddr() net.Addr {
	return &net.TCPAddr{
		Port: s.port,
	}
}

func (s *reuseConn) Close() error {
	err := s.Conn.Close()
	s.reuse.conn = nil
	return err
}
