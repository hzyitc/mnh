package routerForward

import (
	"gitlab.com/NebulousLabs/go-upnp"
)

type UPNP struct {
	udp  bool
	port uint16

	igd *upnp.IGD
}

func newUPNP(udp bool, port int) (Interface, error) {
	igd, err := upnp.Discover()
	if err != nil {
		return nil, err
	}

	u := &UPNP{
		udp,
		uint16(port),

		igd,
	}

	err = u.Redo()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func NewTCPForward_UPNP(port int) (Interface, error) {
	return newUPNP(false, port)
}

func NewUDPForward_UPNP(port int) (Interface, error) {
	return newUPNP(true, port)
}

func (u *UPNP) Redo() error {
	if u.udp {
		return u.igd.ForwardUDP(u.port, "mnh")
	} else {
		return u.igd.ForwardTCP(u.port, "mnh")
	}
}

func (u *UPNP) Close() {
	u.igd.Clear(u.port)
}
