
# mnh
[![GitHub release](https://img.shields.io/github/v/tag/hzyitc/mnh?label=release)](https://github.com/hzyitc/mnh/releases)

[README](README.md) | [中文文档](README_zh.md)

**!!! 注意: `mnh` 仍在开发中, `API` 和 `命令行选项` 可能不向下兼容 !!!**

## 介绍

`mnh`是一个让其他人可以直接访问被NAT服务器的打洞工具。

`mnh`会输出一个可以直接访问的`IP:端口`对。

```
-----------------------
| 协助打洞服务器（有公网）|  <----------（查询IP:端口对）------------
-----------------------                                          |
    ^                                                            |
    |                                                            |
    |   -------------（使用某种方法发送IP:端口对）-------------    |
    |   |                                                   | 或 |
    |   |                                                   ↓    |
---------------                 ~~~~~~~~~~                -----------
| 服务（NAT后）|   <--------     { 互联网 }   <------- --  | 普通用户 |
---------------                 ~~~~~~~~~~                -----------
```

## 使用指南

### 要求

1. 你的网络必须是Full-cone型NAT。
   > 如果你不知道这句话是什么意思，尽管试试这个程序。

2. 如果你的服务在防火墙或者家用路由后面，那么你需要在路由上开启[UPnP](https://en.wikipedia.org/wiki/Universal_Plug_and_Play)或者设置[端口转发](https://en.wikipedia.org/wiki/Port_forwarding)，因为大部分家用路由会阻止入站连接。

### 搭建协助打洞服务器

请转到[mnh_server](https://github.com/hzyitc/mnh_server)。

> `mnh`现在只支持`mnhv1`，但将会在未来支持更多的协议，如[STUN](https://en.wikipedia.org/wiki/STUN)。

### 运行mnh

```
Usage:
  mnh {tcp|udp} --server <server> [flags]

Flags:
  -s, --server string    协助打洞服务器地址(举例 "server.com", "server.com:6641")
                           如果仅指定主机名，它会尝试SRV解析
                           如果SRV解析失败，它会使用默认端口号(6641)
  -i, --id string        一个用来标识你设备的唯一ID

  -m, --mode string      运行模式.
                           TCP支持: demoWeb proxy (默认为 "demoWeb")
                           UDP支持: demoEcho proxy (默认为 "demoEcho")
  -p, --port int         本地洞端口，入口流量将会访问这个端口
  -t, --service string   目标服务地址. 仅proxy模式需要 (默认为 "127.0.0.1:80")

  -r, --routerForward    一个逗号分隔的列表。会按照顺序使用这些来请求路由器进行端口转发 (默认为 "upnp,notice")
                           upnp: UPnP协议
                           notice: 提示你手动设置端口转发
                           none: 手动设置端口转发

  -x, --event-hook       在事件触发后执行命令:
                           转义符:
                             %%: 百分号
                             %e: 事件: connecting fail success disconnected
                             %m: 错误消息
                             %p: 本地洞端口
                             %a: 洞地址

  -h, --help             输出本帮助
```

运行Web服务器进行测试:

```
./mnh tcp --server server.com --id test
```

运行UDP回显服务器进行测试:

```
./mnh udp --server server.com --id udpEcho --mode demoEcho
```

暴露本地Web服务器:

```
./mnh tcp --server server.com --id web --mode proxy --service 127.0.0.1:80
```

`mnh` 默认会尝试使用`upnp`协议来请求路由器进行端口转发。

如果失败了， 将会显示一个提示信息(`notice`)。

可以通过设置 `--routerForward none` 来关闭这两个功能，但请确保你已经正确地设置了端口转发。
(参见 [要求](#要求))

```
./mnh tcp --server server.com --id web --mode proxy --service 127.0.0.1:80 --port 8888 --routerForward none
```
