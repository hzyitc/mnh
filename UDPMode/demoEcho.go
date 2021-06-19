package UDPMode

import (
	"fmt"
	"net"
	"sync"

	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/mnh/routerPortForward"
)

type demoEcho struct {
	port int

	worker      *sync.WaitGroup
	closingChan chan struct{}
	closedChan  chan struct{}

	listener Listener
}

func (s *demoEcho) server_main() {
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
		buf = buf[:n]
		log.Info(fmt.Sprintf("recv %dB from %s: %s", n, addr.String(), string(buf)))
		s.listener.WriteTo(buf, addr)
	}
}

func NewDemoEcho(rpfc routerPortForward.Config, port int) (Interface, error) {
	listener, err := NewListener(rpfc, port)
	if err != nil {
		return nil, err
	}

	s := &demoEcho{
		port,

		new(sync.WaitGroup),
		make(chan struct{}),
		make(chan struct{}),

		listener,
	}

	go s.server_main()

	return s, nil
}

func (s *demoEcho) Dial(addr string) (net.Conn, error) {
	return s.listener.Dial(addr)
}

func (s *demoEcho) ClosedChan() <-chan struct{} {
	return s.closedChan
}

func (s *demoEcho) Close() error {
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

func (s *demoEcho) LocalServiceAddr() net.Addr {
	return s.listener.LocalServiceAddr()
}
