# wovenet

[![Release][1]][2] [![MIT licensed][3]][4]

[1]: https://img.shields.io/github/v/release/kungze/wovenet?color=orange
[2]: https://github.com/kungze/wovenet/releases/latest
[3]: https://img.shields.io/github/license/kungze/wovenet
[4]: LICENSE

[English](./README.md)

应用层网络 VPN，它可以连接不同的私有局域网， 构建多站点的 mesh 网络，提升网络带宽，增强网络稳定性，安全性。

![wovenet topology](./wovenet.png)

直观来看，`wovenet` 就是一个 site-to-site VPN 项目，与 [IPSec VPN](https://en.wikipedia.org/wiki/IPsec)，[wireguard](https://www.wireguard.com/) 类似。这其中一个核心的差异：[IPSec VPN](https://en.wikipedia.org/wiki/IPsec)，[wireguard](https://www.wireguard.com/) 这些通用 VPN 一般是基于三层报文的封装实现，可以称为 L3 layer VPN，而 `wovenet` 是直接封装应用层数据，因此 `wovenet` 可以被成为 **应用层 VPN** （application layer VPN）。相比于这些通用 VPN，`wovenet` 有两大优势：

* `wovenet` 隧道内直接传输应用层数据，不涉及额外包头封装，带宽利用率更高；
* `wovenet` 可以做到应用级别的访问控制，可以控制本站点内哪些应用可以被其他站点访问；

`wovenet` 可以使用户通过本地站点内的套接字访问远端站点的服务（通常位于局域网内）。如上图：在局域网 LAN1 存在 app1， 正常情况下它只能被 LAN1 内的用户访问。但现在，在局域网 LAN3 有用户也想访问 app1。解决方案：只需只需要在两个局域网都启动 `wovenet`（做好正确的配置），在 LAN3 `wovenet` 会打开一个套接字端口，LAN3 的用户就可以通过本局域网内 `wovenet` 打开的套接字端口访问到位于 LAN1 的 APP1。

## 示例

下面是一些示例，展示 `wovenet` 的各种应用场景，希望能通过这些示例快速启发读者，快速确定 `wovenet` 是否值得进一步的研究。**PS：** 我相信下面示例列表绝非 `wovenet` 的全部应用场景，如果各位读者有新的灵感，希望您能向其他读者和作者分享。

在使用过程中有不懂的配置项，请查阅[config.yaml](./config.yaml)

* [释放 VPS 的公网 IP，减少公有云花费](./examples/release-public-ip/README_zh.md)
* [借助 wovenet 实现内网穿透](./examples/reverse-proxy/README_zh.md)
* [借助 wovenet 提升网络传输性能](./examples/network-preformance/README_zh.md)
* [借助 wovenet 应对存在大量动态变化的端口的系统（如：VDI 系统）](./examples/multiple-port/README_zh.md)
* [unix socket 应用转为 tcp socket 应用](./examples/convert-unix-to-tcp/README_zh.md)

在运行这些示例之前，需要先从 [release](https://github.com/kungze/wovenet/releases) 下载最新版本 `wovenet`，并解压。

## 后续功能规划

* 增加 restful 接口，支持动态配置 app
* 增加一套 web ui，降低配置复杂度
* 实现通过打洞的方式建立隧道，摆脱对公网 IP 的依赖
* 增加流量监控功能
* 实现通过 sctp 协议建立隧道，进而增加对 TCP 协议的支持（QUIC 是基于 UDP 的）
