# Converting Unix Socket applications to TCP Socket applications

[简体中文](./README_zh.md)

In real-world environments, many services listen on Unix sockets and can only be accessed from the same machine where the services are running. This creates a lot of inconvenience for developers. `wovenet` can solve this problem by exposing applications that use Unix sockets to remote locations via TCP sockets.

## Environment Information

Launch wovenet instances on two different LAN nodes: `docker-server` and `docker-client`.

* `docker-server` has a public IP address (although it changes dynamically) and runs a Docker service that listens on a Unix socket: /var/run/docker.sock.
* `docker-client` does not have a public IP address and has installed the Docker Python client via pip install docker.

## Wovenet Configuration

### docker-server Configuration

Create a `config.yaml` file with the following content:

```yaml
siteName: docker-server

crypto:
  # This key is used to encrypt sensitive information. Each wovenet instance must be configured with the same value.
  # Important: Do not use the example value provided here, otherwise sensitive information may be leaked.
  key: "3iop#Mg6732ftg#4(ER"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # Important: It is strongly recommended to change the topic. Instances of wovenet with the same topic will form a mesh network.
    topic: "kungze/wovenet/unix-to-tcp-67bhgerAB"

tunnel:
  localSockets:
  - mode: port-forwarding
    transportProtocol: quic
    # Since the public IP is not static, dynamic public IP detection is configured here.
    publicAddress: autoHttpDetect
    httpDetector:
      url: https://ipinfo.io/ip
    publicPort: 60234
    listenPort: 60234

localExposedApps:
- appsocket: unix:/var/run/docker.sock
  appName: docker
```

Start `wovenet`:

```
./wovenet run --config ./config.yaml
```

### docker-client Configuration

Create a `config.yaml` file with the following content:

```yaml
siteName: docker-client

crypto:
  # This key is used to encrypt sensitive information. Each wovenet instance must be configured with the same value.
  # Important: Do not use the example value provided here, otherwise sensitive information may be leaked.
  key: "3iop#Mg6732ftg#4(ER"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # Important: It is strongly recommended to change the topic. Instances of wovenet with the same topic will form a mesh network.
    topic: "kungze/wovenet/unix-to-tcp-67bhgerAB"

# The remote Docker application originally listened on a Unix socket, but now it is accessible via a TCP socket.
remoteApps:
- appName: docker
  siteName: docker-server
  localSocket: tcp:172.26.142.105:38902
```

Start `wovenet`:

```
./wovenet run --config ./config.yaml
```

### Verification

On the `docker-client` side:

```
$ python3
Python 3.12.3 (main, Feb  4 2025, 14:48:35) [GCC 13.3.0] on linux
Type "help", "copyright", "credits" or "license" for more information.
>>> import docker
>>> client = docker.DockerClient(base_url='tcp://172.26.142.105:38902')
>>> for net in client.networks.list():
...     print(net.id, net.name)
...
3f003f373632f58fac1b5d91ae773f929590708881784b209506da6758425538 bridge
84e9e909350ddb30c0c7844ee4e4465226a38eb3c156f863f927d3b8cf6d07fe none
d970d06872ac97f66352a2cc3ac49001df73b57f994f39a84b649fbc6c345968 host
>>>
```
