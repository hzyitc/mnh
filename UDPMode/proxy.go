package UDPMode

import (
	"net"
	"sync"
	"time"

	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/mnh/routerPortForward"
)

type proxy struct {
	port    int
	service net.UDPAddr

	worker      *sync.WaitGroup
	closingChan chan struct{}
	closedChan  chan struct{}

	listener Listener
}

func (s *proxy) server_handle(timeout time.Duration, addr net.Addr, buf []byte) {
	s.worker.Add(+1)
	defer s.worker.Done()

	closing := make(chan int)
	defer close(closing)

	conn, err := s.listener.Add(addr.String())
	if err != nil {
		return
	}
	defer conn.Close()

	log.Info("new connection", conn.RemoteAddr().String())
	c, err := net.DialUDP("udp", nil, &s.service)
	if err != nil {
		return
	}
	defer c.Close()

	copy := func(dst net.Conn, src net.Conn) {
		buf := make([]byte, bufSize)
		for {
			src.SetReadDeadline(time.Now().Add(timeout))
			n, err := src.Read(buf)
			if err != nil {
				break
			}

			dst.SetWriteDeadline(time.Now().Add(timeout))
			_, err = dst.Write(buf[:n])
			if err != nil {
				break
			}
		}
	}

	c.Write(buf)

	go func() {
		copy(conn, c)
		conn.Close()
		closing <- 1
	}()

	go func() {
		copy(c, conn)
		c.Close()
		closing <- 1
	}()

	running := 2
	for {
		select {
		case <-s.closingChan:
			return
		case <-closing:
			running--
			if running == 0 {
				return
			}
		}
	}
}

func (s *proxy) server_main(timeout time.Duration) {
	s.worker.Add(+1)
	defer s.worker.Done()

	for {
		buf := make([]byte, bufSize)
		n, addr, err := s.listener.ReadFrom(buf)
		if err != nil {
			select {
			case <-s.closingChan:
				return
			default:
				log.Error("server_main error", err.Error())
				continue
			}
		}
		go s.server_handle(timeout, addr, buf[:n])
	}
}

func NewProxy(rpfc routerPortForward.Config, port int, service string) (Interface, error) {
	service_addr, err := net.ResolveUDPAddr("udp", service)
	if err != nil {
		return nil, err
	}

	listener, err := NewListener(rpfc, port)
	if err != nil {
		return nil, err
	}

	s := &proxy{
		port,
		*service_addr,

		new(sync.WaitGroup),
		make(chan struct{}),
		make(chan struct{}),

		listener,
	}

	go s.server_main(time.Second * 120)

	return s, nil
}

func (s *proxy) Dial(addr string) (net.Conn, error) {
	return s.listener.Dial(addr)
}

func (s *proxy) ClosedChan() <-chan struct{} {
	return s.closedChan
}

func (s *proxy) Close() error {
	select {
	case <-s.closingChan:
		return nil
	default:
		break
	}
	close(s.closingChan)

	err := s.listener.Close()

	s.worker.Wait()

	close(s.closedChan)
	return err
}

func (s *proxy) LocalServiceAddr() net.Addr {
	return &s.service
}