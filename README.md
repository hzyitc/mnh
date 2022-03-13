# mnh
[![GitHub release](https://img.shields.io/github/v/tag/hzyitc/mnh?label=release)](https://github.com/hzyitc/mnh/releases)

[README](README.md) | [中文文档](README_zh.md)

**!!! Note: `mnh` is currently in development, the `APIs` and `command line options` may not be backward compatible !!!**

## Introduction

`mnh` is a tool that makes exposing a port behind NAT possible.

`mnh` will produce an `ip:port` pair for your NATed server which can be used for public access.

```
--------------------------------
| Help Server (NOT behind NAT) |  <------(query for ip:port pair)-------------
--------------------------------                                             |
        ^                                                                    |
        |                                                                    |
        |       ---------(Use some way to send ip:port pair)-------------    |
        |       |                                                       | or |
        |       |                                                       ↓    |
------------------------                  ~~~~~~~~~~~~                ----------
| Service (behind NAT) |    <--------     { Internet }   <----------  | Guests |
------------------------                  ~~~~~~~~~~~~                ----------
```

## Usage

### Pre-requests

* Your server's network type should be [Full-cone NAT](https://en.wikipedia.org/wiki/Network_address_translation#Methods_of_translation).
  > If you don't known what it means, no worries, just continue.

* If your server is behind a firewall or a household router, you probably need to enable [UPnP](https://en.wikipedia.org/wiki/Universal_Plug_and_Play) or do a [port forwarding](https://en.wikipedia.org/wiki/Port_forwarding) to your server on your router since they may block all incoming traffics.

### Setup up a Help server

Please check [mnh_server](https://github.com/hzyitc/mnh_server).

> Only `mnhv1` protocol is currently supported, more protocols will be added in the future, such as [STUN](https://en.wikipedia.org/wiki/STUN).

### Run mnh

```
Usage:
  mnh {tcp|udp} --server <server> [flags]

Flags:
  -s, --server string    Help server address(Example: "server.com", "server.com:6641")
                           If only specify hostname, it will try SRV resolve.
                           If SRV failed, it will use default port(6641).
  -i, --id string        A unique id to identify your machine

  -m, --mode string      Run mode.
                           TCP support: demoWeb proxy (default "demoWeb")
                           UDP support: demoEcho proxy (default "demoEcho")
  -p, --port int         The local hole port which incoming traffics access to
  -t, --service string   Target service address. Only need in proxy mode (default "127.0.0.1:80")

  -r, --routerForward    A comma-split list which will be used sequentially to request router to do port forwarding (default: "upnp,notice")
                           upnp: UPnP protocol
                           notice: Will notice you to do port forwarding manually
                           none: Do port forwarding manually

  -x, --event-hook       Execute command when event triggered:
                           escape:
                             %%: percent sign
                             %e: Event: connecting fail success disconnected
                             %m: Error message
                             %p: Local hole port
                             %a: Hole addr

  -h, --help             help for mnh
```

Example:

Run a Web server for test:

```
./mnh tcp --server server.com --id test
```

Run a UDP Echo server for test:

```
./mnh udp --server server.com --id udpEcho --mode demoEcho
```

Expose a local web server:

```
./mnh tcp --server server.com --id web --mode proxy --service 127.0.0.1:80
```

`mnh` will attempt to request router to do port forwarding with `upnp` protocol by default.

If failed, it will show a `notice`. 

You can disable these two functions by set `--routerForward none`, make sure you have set up port forwarding correctly.
(See [Pre-requests](#pre-requests))

```
./mnh tcp --server server.com --id web --mode proxy --service 127.0.0.1:80 --port 8888 --routerForward none
```
