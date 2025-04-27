# An effective solution for systems that involve a large number of dynamic ports (such as: VDI Platform)

[简体中文](./README_zh.md)

Some systems need to expose multiple ports when providing external services, and these ports may change dynamically. An example is a VDI (Virtual Desktop Infrastructure) platform.

VDI, also known as a cloud desktop system, treats each cloud desktop as a virtual machine (VM) within the system. These VMs are accessed remotely by users via RDP or SPICE services. Each VM has its own RDP or SPICE service. Therefore, to allow customers to remotely access a VDI system over the public internet, a large number of ports must be exposed. Considering that cloud desktops can be created, deleted, started, and shut down dynamically, the corresponding ports also change dynamically. How to map these VMs' RDP/SPICE ports to public ports becomes a major challenge.

## Reducing the Number of Exposed Ports and Coping with Port Changes Using Wovenet

Since I didn't have an available VDI system when writing this document, and to make it easier for readers to reproduce the example, I simulated multiple VDI VMs by using the `nc` command to start multiple services listening on different ports.

### Environment Information

#### Data Center Side

Public IP: 115.190.109.199

nc services:

* `nc` Service 1: 10.0.2.128:5900
* `nc` Service 2: 10.0.2.129:5900
* `nc` Service 3: 10.0.2.129:5926

#### Client Side

There are three clients, identified as: client1, client2, and client3.

### Data Center Configuration

Create a `config.yaml` file with the following content:

```yaml
siteName: data-center

crypto:
  # This key is used to encrypt sensitive information.
  # All wovenet instances must use the same value.
  # Important: Do not use the value provided in this example, as it may cause sensitive information leakage.
  key: "YHU8D(!dse4UIdewrd56D"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # Important: It is strongly recommended to modify the topic.
    # Wovenet instances with the same topic will form a mesh network.
    topic: "kungze/wovenet/multiple-port-nywgf67s345od"

tunnel:
  localSockets:
  - mode: dedicated-address
    transportProtocol: quic
    publicAddress: 115.190.109.199
    publicPort: 25890
    listenPort: 25890

localExposedApps:
- appName: nc
  mode: range
  portRange:
  - 5900-5990
  addressRange:
  - 10.0.2.0/24
```

**With this configuration, the data center only needs to expose a single UDP port: 25890.**

###### Start `wovenet`:

```
./wovenet run --config ./config.yaml
```

###### Start `nc` services:

* On 10.0.2.128:

```
nc -l 5900
```

* On 10.0.2.129 statr two `nc` services:

`nc` service one:

```
nc -l 5900
```

`nc` service two:

```
nc -l 5926
```


### Client Configuration

Each of the three clients should create a config.yaml file with the following content:

```yaml
# Modify the siteName for the other two clients to client2 and client3 respectively.
siteName: client1

crypto:
  # This key is used to encrypt sensitive information.
  # All wovenet instances must use the same value.
  # Important: Do not use the value provided in this example, as it may cause sensitive information leakage.
  key: "YHU8D(!dse4UIdewrd56D"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # Important: It is strongly recommended to modify the topic.
    # Wovenet instances with the same topic will form a mesh network.
    topic: "kungze/wovenet/multiple-port-nywgf67s345od"

remoteApps:
- siteName: data-center
  appName: nc
  # The other two clients should modify appSocket respectively to:
  # tcp:10.0.2.129:5900 and tcp:10.0.2.129:5926
  appSocket: tcp:10.0.2.128:5900
  localSocket: tcp:127.0.0.1:15900
```

Start `wovenet`:

Each client should run the following command:

```
./wovenet run --config ./config.yaml
```

### Validation

On client1, client2, and client3, respectively, input the following commands and observe the outputs on the data center side:

* client1

```
$ nc 127.0.0.1 15900
hello, I am client1.
```

* client2

```
$ nc 127.0.0.1 15900
hello, I am client2.
```

* client2
```
$ nc 127.0.0.1 15900
hello, I am client3.
```

On the data center side, the corresponding `nc` services will output:

* On 10.0.2.128, `nc` output:

```
$ nc -l 5900
hello, I am client1.
```

* On 10.0.2.129, first `nc` service output:

```
$ nc -l 5900
hello, I am client2.
```

* On 10.0.2.129, second `nc` service output:

```
$ nc -l 5926
hello, I am client3.
```
