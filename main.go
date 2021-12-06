package main

import (
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/google/shlex"
	"github.com/spf13/cobra"

	"github.com/hzyitc/mnh/TCPMode"
	"github.com/hzyitc/mnh/TCPProtocol"
	"github.com/hzyitc/mnh/UDPMode"
	"github.com/hzyitc/mnh/UDPProtocol"
	"github.com/hzyitc/mnh/log"
	"github.com/hzyitc/mnh/routerPortForward"
)

var rootCmd = &cobra.Command{
	Use:   "mnh",
	Short: "A NAT hole punching tool that allows peers directly connect to your NATed server without client.",
	Long: "mnh is a tool that makes exposing a port behind NAT possible.\n" +
		"mnh client will produce an ip:port pair for your NATed server which can be used for public access.",
}

var tcpCmd = &cobra.Command{
	Use: "tcp",
	Run: func(cmd *cobra.Command, args []string) {
		tcp()
	},
}

var udpCmd = &cobra.Command{
	Use: "udp",
	Run: func(cmd *cobra.Command, args []string) {
		udp()
	},
}

var (
	server string
	id     string

	tcpMode string
	udpMode string
	port    int
	service string

	upnpD bool

	eventHook string
)

func commonCmdRegister(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVarP(&server, "server", "s", "server.com", "Help server address (Support SRV) (Default port: 6641)")
	cmd.MarkPersistentFlagRequired("server")
	cmd.PersistentFlags().StringVarP(&id, "id", "i", "", "A unique id to identify your machine")

	cmd.PersistentFlags().IntVarP(&port, "port", "p", 0, "The local hole port which incoming traffics access to")
	cmd.PersistentFlags().StringVarP(&service, "service", "t", "127.0.0.1:80", "Target service address. Only need in proxy mode")

	cmd.PersistentFlags().BoolVarP(&upnpD, "disable-upnp", "u", false, "Disable UPnP")

	cmd.PersistentFlags().StringVarP(&eventHook, "event-hook", "x", "", "Execute command when event triggered")
}

func tcpCmdRegister(cmd *cobra.Command) {
	commonCmdRegister(tcpCmd)
	tcpCmd.PersistentFlags().StringVarP(&tcpMode, "mode", "m", "demoWeb", "Run mode. Available value: demoWeb proxy")

	cmd.AddCommand(tcpCmd)
}

func udpCmdRegister(cmd *cobra.Command) {
	commonCmdRegister(udpCmd)
	udpCmd.PersistentFlags().StringVarP(&udpMode, "mode", "m", "demoEcho", "Run mode. Available value: demoEcho proxy")

	cmd.AddCommand(udpCmd)
}

func runHook(event string, errmsg string, port string, addr string) {
	if eventHook == "" {
		return
	}

	cmdline := strings.NewReplacer(
		"%%", "%",
		"%e", event,
		"%m", errmsg,
		"%p", port,
		"%a", addr,
	).Replace(eventHook)
	log.Debug("Running hook:", cmdline)

	args, err := shlex.Split(cmdline)
	if err != nil {
		log.Error("Split hook error:", err.Error())
		return
	}

	cmd := exec.Command(args[0], args[1:]...)
	err = cmd.Start()
	if err != nil {
		log.Error("Run hook error:", err.Error())
		return
	}

	go func() {
		err = cmd.Wait()
		if err != nil {
			log.Error("Wait hook error:", err.Error())
		}
	}()
}

func main() {
	tcpCmdRegister(rootCmd)
	udpCmdRegister(rootCmd)

	if err := rootCmd.Execute(); err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func tcp() {
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
	switch tcpMode {
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
		log.Error("Unknown mode: ", tcpMode)
		return
	}
	defer mode.Close()

	_, port, _ := net.SplitHostPort(mode.LocalHoleAddr().String())
	for {
		func() {
			runHook("connecting", "", string(port), "")

			protocol, err := TCPProtocol.NewMnhv1(mode, server, id)
			if err != nil {
				log.Error("NewMnhv1 error:", err.Error())
				runHook("fail", err.Error(), "", "")
				return
			}
			defer protocol.Close()

			log.Info("ServiceAddr", mode.ServiceAddr().String())
			log.Info("ServerAddr", protocol.ServerAddr().String())
			log.Info("NATedAddr", protocol.NATedAddr().String())
			log.Info("LocalHoleAddr", protocol.LocalHoleAddr().String())

			log.Info("\n\nNow you can use " + protocol.NATedAddr().String() + " to access your service")

			addr := protocol.NATedAddr().String()
			runHook("success", "", port, addr)
			defer runHook("disconnected", "", port, addr)

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

func udp() {
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
		mode UDPMode.Interface
		err  error
	)
	switch udpMode {
	case "demoEcho":
		mode, err = UDPMode.NewDemoEcho(upnpConfig, port)
		if err != nil {
			log.Error("NewDemoEcho error:", err.Error())
			return
		}
	case "proxy":
		mode, err = UDPMode.NewProxy(upnpConfig, port, service)
		if err != nil {
			log.Error("NewProxy error:", err.Error())
			return
		}
	default:
		log.Error("Unknown mode: ", udpMode)
		return
	}
	defer mode.Close()

	_, port, _ := net.SplitHostPort(mode.LocalHoleAddr().String())
	for {
		func() {
			runHook("connecting", "", string(port), "")

			protocol, err := UDPProtocol.NewMnhv1(mode, server, id)
			if err != nil {
				log.Error("NewMnhv1 error:", err.Error())
				runHook("fail", err.Error(), "", "")
				return
			}
			defer protocol.Close()

			log.Info("ServiceAddr", mode.ServiceAddr().String())
			log.Info("ServerAddr", protocol.ServerAddr().String())
			log.Info("NATedAddr", protocol.NATedAddr().String())
			log.Info("LocalHoleAddr", protocol.LocalHoleAddr().String())

			log.Info("\n\nNow you can use " + protocol.NATedAddr().String() + " to access your service")

			addr := protocol.NATedAddr().String()
			runHook("success", "", port, addr)
			defer runHook("disconnected", "", port, addr)

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
