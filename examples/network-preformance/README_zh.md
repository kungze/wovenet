# 借助 wovenet 提升网络传输性能

[English](./README.md)

有些网络环境，特别是远距离互联网环境会存在丢包现象。对于 TCP 应用，丢包会导致数据包重传，这会极大的影响 TCP 应用的网络性能。

`wovenet` 隧道传输协议默认采用 [QUIC](https://baike.baidu.com/item/%E5%BF%AB%E9%80%9FUDP%E7%BD%91%E7%BB%9C%E8%BF%9E%E6%8E%A5/22785443) 协议。QUIC 有着优秀的重传算法和拥塞控制算法，在存在丢包的网络环境中，QUIC 的表现要远远优于 TCP。

`wovenet` 会把基于 TCP 的应用转为 QUIC 应用在互联网传输，到底目的地后再转为 TCP 应用。因此 `wovenet` 能极大提高不健康网络环境的网络传输性能。

## 测试环境信息

本次测试我准备了两台测试机器： 一台位于美国俄亥俄州（Ohio, America），公网 IP 为 18.119.29.39；另一台位于中国天津（TianJin, China），没有公网 IP。两地地理位置非常遥远。


* 位于俄亥俄州（Ohio）的机器启动了 `iperf` 服务端:

```
$ iperf3 -s
-----------------------------------------------------------
Server listening on 5201 (test #1)
-----------------------------------------------------------
```

* 位于中国天津（TianJin）的机器使用 `iperf` 直接连接到俄亥俄州（Ohio）机器的公网 IP 进行测试，下面是测试结果：

```
$ iperf3 -c 18.119.29.39
Connecting to host 18.119.29.39, port 5201
[  5] local 172.26.142.105 port 54760 connected to 18.119.29.39 port 5201
[ ID] Interval           Transfer     Bitrate         Retr  Cwnd
[  5]   0.00-1.00   sec   640 KBytes  5.24 Mbits/sec    0    154 KBytes
[  5]   1.00-2.00   sec  7.62 MBytes  63.9 Mbits/sec    1   1.52 MBytes
[  5]   2.00-3.00   sec  7.00 MBytes  58.7 Mbits/sec    0   1.71 MBytes
[  5]   3.00-4.00   sec  8.25 MBytes  69.2 Mbits/sec    0   1.85 MBytes
[  5]   4.00-5.00   sec  9.50 MBytes  79.7 Mbits/sec    0   1.95 MBytes
[  5]   5.00-6.00   sec  9.62 MBytes  80.7 Mbits/sec    0   2.04 MBytes
[  5]   6.00-7.00   sec  9.62 MBytes  80.7 Mbits/sec    0   2.09 MBytes
[  5]   7.00-8.00   sec  10.2 MBytes  86.1 Mbits/sec    4   1.48 MBytes
[  5]   8.00-9.00   sec  4.88 MBytes  40.9 Mbits/sec   48    790 KBytes
[  5]   9.00-10.00  sec  4.12 MBytes  34.6 Mbits/sec    0    853 KBytes
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-10.00  sec  71.5 MBytes  60.0 Mbits/sec   53             sender
[  5]   0.00-10.21  sec  69.1 MBytes  56.8 Mbits/sec                  receiver

iperf Done.
```

## 使用 wovenet 对网络进行优化

### Ohio 机器配置

创建 `config.yaml` 文件，内容如下：

```yaml
siteName: Ohio

crypto:
  # 这个 key 是用来加密敏感信息的，要求每个 wovenet 实例都要配置相同的值。
  # 特别注意：不要使用本示例提供的值，否则可能造成敏感信息泄露。
  key: "06Uw12TYdYUIEddse#r"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # 特别注意：强烈建议修改 topic，相同 topic 的 wovenet 实例会共同组成一个 mesh 网络。
    topic: "kungze/wovenet/network-performance-Po783tfdx"

tunnel:
  localSockets:
  - mode: dedicated-address
    transportProtocol: quic
    publicAddress: 18.119.29.39
    publicPort: 22892
    listenPort: 22892

localExposedApps:
- appSocket: tcp:127.0.0.1:5201
  appName: iperf
```

启动 `wovenet`：

```
./wovenet run --config ./config.yaml
```

**注意：** 需要在安全组/防火墙上设置规程，放行 UDP 22892 端口。

### TianJin 机器配置

创建 `config.yaml` 文件，内容如下：

```yaml
siteName: TianJin

crypto:
  # 这个 key 是用来加密敏感信息的，要求每个 wovenet 实例都要配置相同的值。
  # 特别注意：不要使用本示例提供的值，否则可能造成敏感信息泄露。
  key: "06Uw12TYdYUIEddse#r"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # 特别注意：强烈建议修改 topic，相同 topic 的 wovenet 实例会共同组成一个 mesh 网络。
    topic: "kungze/wovenet/network-performance-Po783tfdx"

remoteApps:
- siteName: Ohio
  appName: iperf
  localSocket: tcp:127.0.0.1:5201
```

启动 `wovenet`：

```
./wovenet run --config ./config.yaml
```

### 启动 iperf 通过 wovenet 进行带宽测试

```
$ iperf3 -c 127.0.0.1
Connecting to host 127.0.0.1, port 5201
[  5] local 127.0.0.1 port 52664 connected to 127.0.0.1 port 5201
[ ID] Interval           Transfer     Bitrate         Retr  Cwnd
[  5]   0.00-1.00   sec  5.50 MBytes  46.1 Mbits/sec    1   1.37 MBytes
[  5]   1.00-2.00   sec  14.1 MBytes   118 Mbits/sec    0   1.37 MBytes
[  5]   2.00-3.00   sec  21.9 MBytes   184 Mbits/sec    1   1.37 MBytes
[  5]   3.00-4.00   sec  18.9 MBytes   158 Mbits/sec    1   1.06 MBytes
[  5]   4.00-5.00   sec  22.1 MBytes   186 Mbits/sec    0   1.06 MBytes
[  5]   5.00-6.00   sec  22.0 MBytes   185 Mbits/sec    0   1.06 MBytes
[  5]   6.00-7.00   sec  20.9 MBytes   175 Mbits/sec    1   1.06 MBytes
[  5]   7.00-8.00   sec  22.9 MBytes   192 Mbits/sec    1   1.06 MBytes
[  5]   8.00-9.00   sec  19.9 MBytes   167 Mbits/sec    1   1.06 MBytes
[  5]   9.00-10.00  sec  10.6 MBytes  89.1 Mbits/sec    0   1.06 MBytes
- - - - - - - - - - - - - - - - - - - - - - - - -
[ ID] Interval           Transfer     Bitrate         Retr
[  5]   0.00-10.00  sec   179 MBytes   150 Mbits/sec    6             sender
[  5]   0.00-10.21  sec   174 MBytes   143 Mbits/sec                  receiver

iperf Done.
```

从测试结果可以看到，带宽有了大幅提升。**注意：** `wovenet` 只有在网络不太健康的环境中才会有如此优异的表现，在健康网络环境中对网络带宽几乎没有影响，参考[释放 VPS 的公网 IP，减少公有云花费](../release-public-ip/README_zh.md) 测试结果。

