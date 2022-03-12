package routerForward

import "github.com/hzyitc/mnh/log"

type Notice struct {
}

func NewTCPForward_Notice(port int) (Interface, error) {
	return &Notice{}, nil
}

func NewUDPForward_Notice(port int) (Interface, error) {
	return &Notice{}, nil
}

func (u *Notice) Redo() error {
	log.Info("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	log.Info("!!!Notice: please set up port forwarding correctly!!!")
	log.Info("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!")
	return nil
}

func (u *Notice) Close() {
}
