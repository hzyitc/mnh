package routerPortForward

import (
	"github.com/hzyitc/mnh/log"
	"gitlab.com/NebulousLabs/go-upnp"
)

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
		goto NO_UPNP
	} else {
		log.Info("Attempting UPnP port forward, this might take a while...")
		log.Info("You can disable this behavior by adding --disable-upnp")
		d, err := upnp.Discover()
		if err != nil {
			// Do not quit if upnp fails, fix #1
			log.Info("UPnP port forward failed.")
			log.Info("Falling back to non UPnP mode, you might need to do a port forward manually on your router.")
			goto NO_UPNP
		}

		d.Forward(p, "mnh")
		return &upnpImpl{
			config,
			p,
			d,
		}, nil
	}

NO_UPNP:
	return &upnpImpl{
		config,
		p,
		nil,
	}, nil
}

func (s *upnpImpl) Close() {
	if s.config.Enable {
		s.d.Clear(s.port)
	}
}
