package routerPortForward

import "gitlab.com/NebulousLabs/go-upnp"

type Config struct {
	Enable bool
}

type Interface interface {
	Close()
}

type upnpImpl struct {
	config Config
	port   uint16

	d *upnp.IGD
}

func New(config Config, port int) (Interface, error) {
	p := uint16(port)

	if !config.Enable {
		return &upnpImpl{
			config,
			p,

			nil,
		}, nil
	}

	d, err := upnp.Discover()
	if err != nil {
		return nil, err
	}

	d.Forward(p, "mnh")

	return &upnpImpl{
		config,
		p,

		d,
	}, nil
}

func (s *upnpImpl) Close() {
	if s.config.Enable {
		s.d.Clear(s.port)
	}
}
