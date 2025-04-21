# wovenet

[简体中文](./README_zh.md)

Link separate private network sites, build a site-to-site network. Improve network stability, security and traffic performance.

![wovenet topology](./wovenet.png)

From an intuitive perspective, this project appears to be an implementation of an IPSec VPN. While they provide similar functionality, there are many differences between them:

* IPSec VPN is an Ethernet-level VPN that transmits Ethernet frames through the tunnel, whereas `wovenet` is an application-layer VPN that transmits application-level data through the tunnel.

  `wovenet` offers two key advantages:
  * First, `wovenet` directly transmits application-layer data within the tunnel without adding extra packet headers, resulting in better bandwidth efficiency.
  * Second, `wovenet` can control which applications on the local site are accessible to other sites. IPSec VPN cannot do this (once a tunnel is established, both sites can access all services on all hosts within each other’s internal network).

* IPSec VPN requires each site to have a fixed public IP address, but `wovenet` does not.

  In its current version, `wovenet` only requires some sites to have a fixed public IP address. For example, in the topology shown above, only any two sites need to have fixed public IPs. If only two sites need to be connected, then only one of them needs a fixed public IP.

  Future versions will include features like `dynamic public IP detection` and `NAT traversal`.
  * `Dynamic public IP detection` is used to handle cases where the public IP changes due to certain reasons (e.g., manual changes such as floating IPs or elastic IPs in cloud environments; or IP changes caused by DHCP lease expiration, as commonly happens with home broadband routers).
  * `NAT traversal` is designed for scenarios where sites cannot configure public IP mappings on their NAT gateways (i.e., port forwarding is not possible).

* `wovenet` supports dual-stack IPv4 and IPv6 operation, which not only provides redundant links for tunnels but also enables load balancing across links to increase overall tunnel bandwidth.

## User Cases

* [Dual-Stack Configuration with IPv4 and IPv6 to Enhance Stability and Cut Costs](./examples/v4-v6-dual-stack/README.md)
* [Internal Network Penetration with wovenet](./examples/reverse-proxy/README.md)

If you encounter any configuration options you don't understand during use, please refer to [config.yaml](./config.yaml).

## TODO LIST

* Implement automatic reconnection mechanism after tunnel interruption
* Add dynamic public IP retrieval to prevent tunnel reconnection failure due to IP changes
* Allow client-side configuration of which apps can access remote sites
* Add RESTful API to support dynamic app configuration
* Develop a web UI to simplify configuration
* Establish tunnels using hole punching to eliminate reliance on public IPs
* Add traffic monitoring functionality
* Supports establishing a tunnel using the STCP protocol, thereby enabling support for the TCP protocol (QUIC is based on UDP)
