package routerForward

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/sync/errgroup"

	"github.com/huin/goupnp"
	"github.com/huin/goupnp/dcps/internetgateway2"
)

type UPNP struct {
	protocol string
	port     uint16

	client interface {
		GetServiceClient() *goupnp.ServiceClient

		AddPortMappingCtx(
			ctx context.Context,
			NewRemoteHost string,
			NewExternalPort uint16,
			NewProtocol string,
			NewInternalPort uint16,
			NewInternalClient string,
			NewEnabled bool,
			NewPortMappingDescription string,
			NewLeaseDuration uint32,
		) (err error)

		DeletePortMappingCtx(
			ctx context.Context,
			NewRemoteHost string,
			NewExternalPort uint16,
			NewProtocol string,
		) (err error)
	}
}

func newUPNP(protocol string, port int) (Interface, error) {
	u := &UPNP{
		protocol: protocol,
		port:     uint16(port),
	}

	ctx, done := signal.NotifyContext(context.TODO(), os.Interrupt, syscall.SIGTERM)
	defer done()

	tasks, _ := errgroup.WithContext(ctx)

	// Request each type of client in parallel, and return what is found.
	var ip1Clients []*internetgateway2.WANIPConnection1
	tasks.Go(func() error {
		var err error
		ip1Clients, _, err = internetgateway2.NewWANIPConnection1ClientsCtx(ctx)
		return err
	})

	var ip2Clients []*internetgateway2.WANIPConnection2
	tasks.Go(func() error {
		var err error
		ip2Clients, _, err = internetgateway2.NewWANIPConnection2ClientsCtx(ctx)
		return err
	})

	var ppp1Clients []*internetgateway2.WANPPPConnection1
	tasks.Go(func() error {
		var err error
		ppp1Clients, _, err = internetgateway2.NewWANPPPConnection1ClientsCtx(ctx)
		return err
	})

	if err := tasks.Wait(); err != nil {
		return nil, err
	}

	// Trivial handling for where we find exactly one device to talk to, you
	// might want to provide more flexible handling than this if multiple
	// devices are found.
	switch {
	case len(ip2Clients) == 1:
		u.client = ip2Clients[0]
	case len(ip1Clients) == 1:
		u.client = ip1Clients[0]
	case len(ppp1Clients) == 1:
		u.client = ppp1Clients[0]
	default:
		return nil, errors.New("multiple or no services found")
	}

	err := u.Redo()
	if err != nil {
		return nil, err
	}

	return u, nil
}

func NewTCPForward_UPNP(port int) (Interface, error) {
	return newUPNP("TCP", port)
}

func NewUDPForward_UPNP(port int) (Interface, error) {
	return newUPNP("UDP", port)
}

func (u *UPNP) Redo() error {
	return u.client.AddPortMappingCtx(
		context.Background(),
		"",
		u.port,
		u.protocol,
		u.port,
		u.client.GetServiceClient().LocalAddr().String(),
		true,
		"mnh",
		0,
	)
}

func (u *UPNP) Close() {
	u.client.DeletePortMappingCtx(
		context.Background(),
		"",
		u.port,
		u.protocol,
	)
}
