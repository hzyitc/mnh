package TCPMode

import (
	"fmt"
	"net"
	"sync"

	"github.com/hzyitc/mnh/log"
)

type demoWeb struct {
	port int

	worker      *sync.WaitGroup
	closingChan chan struct{}
	closedChan  chan struct{}

	listener Listener
}

func (s *demoWeb) server_handle(conn net.Conn) {
	s.worker.Add(+1)
	defer s.worker.Done()

	defer conn.Close()

	log.Info("new connection", conn.RemoteAddr().String())

	body := "It's working!!!<br />\n" +
		"Your address is " + conn.RemoteAddr().String()

	header := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Server: mnh demoWeb\r\n"+
			"Content-Length: %d\r\n"+
			"\r\n",
		len(body))
	conn.Write([]byte(header + body))
	conn.Close()
}

func (s *demoWeb) server_main() {
	s.worker.Add(+1)
	defer s.worker.Done()

	for {
		conn, err := s.listener.Accept()
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

func NewDemoWeb(rfc string, port int) (Interface, error) {
	listener, err := NewListener(rfc, port)
	if err != nil {
		return nil, err
	}

	s := &demoWeb{
		port,

		new(sync.WaitGroup),
		make(chan struct{}),
		make(chan struct{}),

		listener,
	}

	go s.server_main()

	return s, nil
}

func (s *demoWeb) Dial(addr string) (net.Conn, error) {
	return s.listener.Dial(addr)
}

func (s *demoWeb) ClosedChan() <-chan struct{} {
	return s.closedChan
}

func (s *demoWeb) Close() error {
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

func (s *demoWeb) LocalHoleAddr() net.Addr {
	return s.listener.LocalHoleAddr()
}

func (s *demoWeb) ServiceAddr() net.Addr {
	return s.listener.ServiceAddr()
}
