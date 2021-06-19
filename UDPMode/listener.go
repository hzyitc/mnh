package UDPMode

import (
	"errors"
	"math"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/mnh/routerPortForward"
)

const listener_bufferLen = 64

type Listener interface {
	Add(addr string) (net.Conn, error)

	Interface
	net.PacketConn
}

type listener struct {
	port int
	rpf  routerPortForward.Interface

	worker      *sync.WaitGroup
	closingChan chan struct{}
	closedChan  chan struct{}

	server net.PacketConn

	writeDeadline time.Time
	readDeadline  time.Time

	buffer chan *listenerBuffer
	conns  sync.Map
	conn   net.Conn
}

type listenerBuffer struct {
	addr   net.Addr
	buffer []byte
}

type listenerConn struct {
	listener *listener
	addr     net.Addr

	closed chan interface{}

	writeDeadline time.Time
	readDeadline  time.Time

	buffer chan []byte
}

func (s *listener) server_main() {
	s.worker.Add(1)
	defer s.worker.Done()

	defer s.Close()

	for {
		buf := make([]byte, bufSize)
		n, addr, err := s.server.ReadFrom(buf)
		if err != nil {
			log.Error("ReadFrom error", err.Error())
			return
		}
		buf = buf[:n]

		v, found := s.conns.Load(addr.String())
		if found {
			c := v.(*listenerConn)
			select {
			case c.buffer <- buf:
			default:
			}
		} else {
			select {
			case s.buffer <- &listenerBuffer{addr, buf}:
			default:
			}
		}
	}

}

func NewListener(rpfc routerPortForward.Config, port int) (Listener, error) {
	rpf, err := routerPortForward.New(rpfc, port)
	if err != nil {
		return nil, err
	}

	local := "0.0.0.0:" + strconv.Itoa(port)
	server, err := net.ListenPacket("udp4", local)
	if err != nil {
		return nil, err
	}

	addr, err := net.ResolveUDPAddr("udp", server.LocalAddr().String())
	if err != nil {
		server.Close()
		return nil, err
	}
	port = addr.Port

	s := &listener{
		port,
		rpf,

		new(sync.WaitGroup),
		make(chan struct{}),
		make(chan struct{}),

		server,

		time.Time{},
		time.Time{},

		make(chan *listenerBuffer, listener_bufferLen),
		sync.Map{},
		nil,
	}

	go s.server_main()

	return s, nil
}

func (s *listener) Add(addr string) (net.Conn, error) {
	remote, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		return nil, err
	}

	c := &listenerConn{
		s,
		remote,

		make(chan interface{}),

		time.Time{},
		time.Time{},

		make(chan []byte, listener_bufferLen),
	}

	_, found := s.conns.LoadOrStore(remote.String(), c)
	if found {
		return nil, errors.New("connected")
	}

	return c, nil
}

func (s *listener) Dial(addr string) (net.Conn, error) {
	s.rpf.Redo()

	if s.conn != nil {
		return nil, errors.New("double dial")
	}

	select {
	case <-s.closingChan:
		return nil, net.ErrClosed
	default:
		c, err := s.Add(addr)
		if err != nil {
			return nil, err
		}

		s.conn = c

		return c, nil
	}
}

func (s *listener) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	t := time.Until(s.readDeadline)
	if s.readDeadline.IsZero() {
		t = time.Duration(math.MaxInt64)
	}

	select {
	case <-s.closingChan:
		return 0, nil, net.ErrClosed
	case buf := <-s.buffer:
		return copy(p, buf.buffer), buf.addr, nil
	case <-time.After(t):
		return 0, nil, errors.New("timeout")
	}
}

func (s *listener) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	select {
	case <-s.closingChan:
		return 0, net.ErrClosed
	default:
	}

	s.server.SetWriteDeadline(s.writeDeadline)
	return s.server.WriteTo(p, addr)
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

	err := s.server.Close()
	s.conns.Range(func(key, value interface{}) bool {
		conn := value.(*listenerConn)
		conn.Close()
		return true
	})

	s.worker.Wait()

	close(s.closedChan)
	return err
}

func (s *listener) LocalAddr() net.Addr {
	return s.server.LocalAddr()
}

func (s *listener) LocalServiceAddr() net.Addr {
	return &net.UDPAddr{
		Port: s.port,
	}
}

func (s *listener) SetDeadline(t time.Time) error {
	select {
	case <-s.closingChan:
		return net.ErrClosed
	default:
	}

	s.SetReadDeadline(t)
	s.SetWriteDeadline(t)
	return nil
}

func (s *listener) SetReadDeadline(t time.Time) error {
	select {
	case <-s.closingChan:
		return net.ErrClosed
	default:
	}

	s.readDeadline = t

	return nil
}

func (s *listener) SetWriteDeadline(t time.Time) error {
	select {
	case <-s.closingChan:
		return net.ErrClosed
	default:
	}

	s.writeDeadline = t

	return nil
}

func (s *listenerConn) Read(b []byte) (n int, err error) {
	t := time.Until(s.readDeadline)
	if s.readDeadline.IsZero() {
		t = time.Duration(math.MaxInt64)
	}

	select {
	case <-s.closed:
		return 0, net.ErrClosed
	case buf := <-s.buffer:
		return copy(b, buf), nil
	case <-time.After(t):
		return 0, errors.New("timeout")
	}
}

func (s *listenerConn) Write(b []byte) (n int, err error) {
	select {
	case <-s.closed:
		return 0, net.ErrClosed
	default:
		s.listener.server.SetWriteDeadline(s.writeDeadline)
		return s.listener.WriteTo(b, s.addr)
	}
}

func (s *listenerConn) Close() error {
	select {
	case <-s.closed:
		return net.ErrClosed
	default:
	}

	close(s.closed)
	s.listener.conns.Delete(s.addr.String())

	if s == s.listener.conn {
		s.listener.conn = nil
	}

	return nil
}

func (s *listenerConn) LocalAddr() net.Addr {
	return s.listener.LocalAddr()
}

func (s *listenerConn) RemoteAddr() net.Addr {
	return s.addr
}

func (s *listenerConn) SetDeadline(t time.Time) error {
	select {
	case <-s.closed:
		return net.ErrClosed
	default:
	}

	s.SetReadDeadline(t)
	s.SetWriteDeadline(t)
	return nil
}

func (s *listenerConn) SetReadDeadline(t time.Time) error {
	select {
	case <-s.closed:
		return net.ErrClosed
	default:
	}

	s.readDeadline = t

	return nil
}

func (s *listenerConn) SetWriteDeadline(t time.Time) error {
	select {
	case <-s.closed:
		return net.ErrClosed
	default:
	}

	s.writeDeadline = t

	return nil
}
