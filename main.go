package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/spf13/cobra"

	"github.com/hzyitc/mnh/TCPMode"
	"github.com/hzyitc/mnh/TCPProtocol"
	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/mnh/routerPortForward"
)

var rootCmd = &cobra.Command{
	Use:   "mnh",
	Short: "A NAT hole punching tool that allows peers directly connect to your NATed server without client.",
	Long: "mnh is a tool that makes exposing a port behind NAT possible.\n" +
		"mnh client will produce an ip:port pair for your NATed server which can be used for public access.",
	Run: func(cmd *cobra.Command, args []string) {
		trueMain()
	},
}

var (
	server string
	id     string

	modeS   string
	port    int
	service string

	upnpD bool
)

func cmdRegister(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&server, "server", "s", "server.com:12345", "Help server address")
	cmd.MarkPersistentFlagRequired("server")
	cmd.PersistentFlags().StringVarP(&id, "id", "i", "", "A unique id to identify your machine")

	cmd.PersistentFlags().StringVarP(&modeS, "mode", "m", "demoWeb", "Run mode. Available value: demoWeb proxy")
	cmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "The local hole port which incoming traffics access to")
	cmd.PersistentFlags().StringVarP(&service, "service", "t", "127.0.0.1:80", "Target service address. Only need in proxy mode")

	cmd.PersistentFlags().BoolVarP(&upnpD, "disable-upnp", "u", false, "Disable UPnP")
}

func main() {
	cmdRegister(rootCmd)
	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func trueMain() {
	c := make(chan os.Signal)
	closing := make(chan struct{})
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		close(closing)
	}()

	upnpConfig := routerPortForward.Config{
		Enable: !upnpD,
	}

	var (
		mode TCPMode.Interface
		err  error
	)
	switch modeS {
	case "demoWeb":
		mode, err = TCPMode.NewDemoWeb(upnpConfig, port)
		if err != nil {
			log.Error("NewDemoWeb error:", err.Error())
			return
		}
	case "proxy":
		mode, err = TCPMode.NewProxy(upnpConfig, port, service)
		if err != nil {
			log.Error("NewProxy error:", err.Error())
			return
		}
	default:
		log.Error("Unknown mode: ", modeS)
		return
	}
	defer mode.Close()

	for {
		func() {
			protocol, err := TCPProtocol.NewMnhv1(mode, server, id)
			if err != nil {
				log.Error("NewMnhv1 error:", err.Error())
				return
			}
			defer protocol.Close()

			log.Info("LocalServiceAddr", mode.LocalServiceAddr().String())
			log.Info("RemoteServerAddr", protocol.RemoteServerAddr().String())
			log.Info("RemoteHoleAddr", protocol.RemoteHoleAddr().String())
			log.Info("LocalHoleAddr", protocol.LocalHoleAddr().String())

			log.Info("\n\nNow you can use " + protocol.RemoteHoleAddr().String() + " to access your service")

			select {
			case <-protocol.ClosedChan():
				return
			case <-closing:
				return
			}
		}()

		select {
		case <-closing:
			return
		default:
		}

		time.Sleep(time.Second)
		log.Info("Reconnecting ...")
	}
}
