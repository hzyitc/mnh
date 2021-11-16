package UDPProtocol

import (
	"errors"
	"net"
	"sync"
	"time"

	"github.com/hzyitc/mnh/UDPMode"
	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/netutils"
)

type mnhv1 struct {
	mode UDPMode.Interface

	conn     net.Conn
	lastSeen time.Time

	worker      *sync.WaitGroup
	closingChan chan struct{}
	closedChan  chan struct{}

	holeAddr net.Addr
}

func (s *mnhv1) keepalive(duration time.Duration, timeout time.Duration) {
	defer s.Close()

	s.worker.Add(+1)
	defer s.worker.Done()

	for {
		buf := []byte("heartbeat\n")
		_, err := s.conn.Write(buf)
		if err != nil {
			log.Error("send heartbeat fail:", err.Error())
			return
		}

		t := time.Now()

		s.conn.SetReadDeadline(t.Add(duration))
		buf = make([]byte, 255)
		_, err = s.conn.Read(buf)
		if err == nil {
			s.lastSeen = time.Now()
		} else {
			log.Debug("read heartbeat fail:", err.Error())
		}

		if time.Since(s.lastSeen) > timeout {
			log.Debug("heartbeat timeout:", time.Since(s.lastSeen).String())
			return
		}

		select {
		case <-s.closingChan:
			return
		case <-time.After(duration - time.Since(t)):
		}
	}
}

func NewMnhv1(m UDPMode.Interface, server string, id string) (Interface, error) {
	addr, err := netutils.ResolveAddr("udp", server, "mnhv1", 6641)
	if err != nil {
		return nil, err
	}

	conn, err := m.Dial(addr.String())
	if err != nil {
		return nil, err
	}

	_, err = conn.Write([]byte("mnhv1 " + id + "\n"))
	if err != nil {
		conn.Close()
		return nil, err
	}

	buf := make([]byte, 512)
	conn.SetReadDeadline(time.Now().Add(time.Second * 10))
	n, err := conn.Read(buf)
	if err != nil {
		conn.Close()
		return nil, err
	}
	msg := string(buf[:n])

	holeAddr := netutils.ParseAddr("", msg)
	if err != nil {
		conn.Close()
		return nil, errors.New("fail to parse \"" + msg + "\"")
	}

	s := &mnhv1{
		m,

		conn,
		time.Now(),

		new(sync.WaitGroup),
		make(chan struct{}),
		make(chan struct{}),

		holeAddr,
	}
	go s.keepalive((time.Second * 5), (time.Second * 30))

	return s, nil
}

func (s *mnhv1) ClosedChan() <-chan struct{} {
	return s.closedChan
}

func (s *mnhv1) Close() error {
	select {
	case <-s.closingChan:
		return nil
	default:
		break
	}
	close(s.closingChan)

	err := s.conn.Close()

	s.worker.Wait()

	close(s.closedChan)
	return err
}

func (s *mnhv1) ServerAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *mnhv1) LocalHoleAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *mnhv1) NATedAddr() net.Addr {
	return s.holeAddr
}
