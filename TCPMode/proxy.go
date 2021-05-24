package TCPMode

import (
	"io"
	"net"
	"sync"

	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/mnh/routerPortForward"
)

type proxy struct {
	port    int
	service net.TCPAddr

	worker      *sync.WaitGroup
	closingChan chan struct{}
	closedChan  chan struct{}

	listener Interface
	server   net.Listener
}

func (s *proxy) server_handle(conn net.Conn) {
	s.worker.Add(+1)
	defer s.worker.Done()

	closing := make(chan int)
	defer close(closing)

	defer conn.Close()

	log.Info("new connection", conn.RemoteAddr().String())
	c, err := net.DialTCP("tcp", nil, &s.service)
	if err != nil {
		return
	}
	defer c.Close()

	go func() {
		io.Copy(conn, c)
		conn.Close()
		closing <- 1
	}()

	go func() {
		io.Copy(c, conn)
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

func (s *proxy) server_main() {
	s.worker.Add(+1)
	defer s.worker.Done()

	for {
		conn, err := s.server.Accept()
		if err != nil {
			select {
			case <-s.closingChan:
				return
			default:
				log.Error("server_main error", err.Error())
				continue
			}
		}
		go s.server_handle(conn)
	}
}

func NewProxy(rpfc routerPortForward.Config, port int, service string) (Interface, error) {
	service_addr, err := net.ResolveTCPAddr("tcp", service)
	if err != nil {
		return nil, err
	}

	listener, server, err := NewListener(rpfc, port)
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
		server,
	}

	go s.server_main()

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
