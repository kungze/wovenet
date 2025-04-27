# Improve network performance with wovenet

[简体中文](./README_zh.md)

In some network environments, especially over long-distance internet connections, packet loss may occur. For TCP-based applications, packet loss leads to retransmissions, which can significantly degrade network performance.

The `wovenet` tunnel transmission protocol uses [QUIC](https://en.wikipedia.org/wiki/QUIC) by default. QUIC features excellent retransmission and congestion control algorithms, and performs far better than TCP in environments with packet loss.

`wovenet` converts TCP-based applications into QUIC-based transmissions over the internet and then reverts them back to TCP at the destination. As a result, `wovenet` can significantly improve network transmission performance in poor network conditions.

## Test Environment Information

For this test, I prepared two machines: one located in Ohio, USA, with a public IP address of 18.119.29.39, and the other located in Tianjin, China, without a public IP. The two locations are geographically very far apart.

* The machine in Ohio started an iperf server:

```
$ iperf3 -s
-----------------------------------------------------------
Server listening on 5201 (test #1)
-----------------------------------------------------------
```

* The machine in Tianjin directly connected to the Ohio machine’s public IP using iperf for testing. Here are the results:

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

## Optimizing the Network with wovenet

### Configuration on the Ohio Machine

Create a `config.yaml` file with the following content:

```yaml
siteName: Ohio

crypto:
  # This key is used to encrypt sensitive information.
  # Each wovenet instance must use the same key.
  # WARNING: Do not use the key provided in this example to avoid potential leaks.
  key: "06Uw12TYdYUIEddse#r"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # WARNING: Strongly recommend modifying the topic.
    # Wovenet instances sharing the same topic will form a mesh network.
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

Start `wovenet` with:

```
./wovenet run --config ./config.yaml
```

**Note:**  You must add a firewall/security group rule on the public cloud platform to allow UDP port 22892.

### Configuration on the Tianjin Machine

```yaml
siteName: TianJin

crypto:
  # This key is used to encrypt sensitive information.
  # Each wovenet instance must use the same key.
  # WARNING: Do not use the key provided in this example to avoid potential leaks.
  key: "06Uw12TYdYUIEddse#r"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # WARNING: Strongly recommend modifying the topic.
    # Wovenet instances sharing the same topic will form a mesh network.
    topic: "kungze/wovenet/network-performance-Po783tfdx"

remoteApps:
- siteName: Ohio
  appName: iperf
  localSocket: tcp:127.0.0.1:5201
```

Start `wovenet` with:

```
./wovenet run --config ./config.yaml
```

#### Start Bandwidth Test Using iperf via wovenet

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

From the test results, we can see that the bandwidth has improved significantly.
Note: wovenet delivers outstanding results only in poor network conditions. In healthy networks, it has little to no impact on bandwidth. For more information, see [Release Public IP from VPS to Reduce Public Cloud Costs](../release-public-ip/README.md).
