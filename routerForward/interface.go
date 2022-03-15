package routerForward

import (
	"errors"
	"strings"

	"github.com/hzyitc/mnh/log"
)

type Interface interface {
	Redo() error
	Close()
}

var ProtocolList = []string{"upnp", "notice", "none"}

func findUnsupportedList(support, need []string) []string {
	m := make(map[string]bool)
	for _, p := range support {
		m[p] = true
	}

	s := make([]string, 0)
	for _, p := range need {
		if !m[p] {
			s = append(s, p)
		}
	}

	return s
}

func NewTCPForward(config string, port int) (Interface, error) {
	list := strings.Split(config, ",")

	unsupported := findUnsupportedList(ProtocolList, list)
	if len(unsupported) != 0 {
		return nil, errors.New("Unsupport RouterForward protocol: " + strings.Join(unsupported, " "))
	}

	log.Info("Attempting to request router to do port forwarding")
	for _, p := range list {
		i, err := func() (Interface, error) {
			switch p {
			case "upnp":
				return NewTCPForward_UPNP(port)
			case "notice":
				return NewTCPForward_Notice(port)
			case "none":
				return NewTCPForward_None(port)
			default:
				// Never reach
				return nil, nil
			}
		}()

		if err == nil {
			return i, nil
		}
	}

	return nil, errors.New("Unable to finish RouterForward")
}

func NewUDPForward(config string, port int) (Interface, error) {
	list := strings.Split(config, ",")

	unsupported := findUnsupportedList(ProtocolList, list)
	if len(unsupported) != 0 {
		return nil, errors.New("Unsupport RouterForward protocol: " + strings.Join(unsupported, " "))
	}

	log.Info("Attempting to request router to do port forwarding")
	for _, p := range list {
		i, err := func() (Interface, error) {
			switch p {
			case "upnp":
				return NewUDPForward_UPNP(port)
			case "notice":
				return NewUDPForward_Notice(port)
			case "none":
				return NewUDPForward_None(port)
			default:
				// Never reach
				return nil, nil
			}
		}()

		if err == nil {
			return i, nil
		}
	}

	return nil, errors.New("Unable to finish RouterForward")
}
