# unix socket 应用转为 tcp socket 应用

[English](./README.md)

在实际环境中，许多服务监听在 Unix 套接字上，只能在与这些服务位于同一台机器上访问，这给开发带来了很多不便。`wovenet` 可以解决这一问题，将 Unix Socket 的应用暴露给远端，通过 TCP Socket 访问。

## 环境信息

分别在两个局域网节点上启动 `wovenet` 实例：`docker-server`，`docker-client`。

* `docker-server` 拥有公网 IP（但是公网 IP 是动态变化的），并且启动了一个 docker 服务，监听在 unix 套接字：/var/run/docker.sock
* `docker-client` 没有公网 IP，通过 `pip install docker` 安装了 docker python client。 

## wovenet 配置

### docker-server 配置

创建 `config.yaml` 文件，内容如下：

```yaml
siteName: docker-server

crypto:
  # 这个 key 是用来加密敏感信息的，要求每个 wovenet 实例都要配置相同的值。
  # 特别注意：不要使用本示例提供的值，否则可能造成敏感信息泄露。
  key: "3iop#Mg6732ftg#4(ER"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
     # 特别注意：强烈建议修改 topic，相同 topic 的 wovenet 实例会共同组成一个 mesh 网络。
    topic: "kungze/wovenet/unix-to-tcp-67bhgerAB"

tunnel:
  localSockets:
  - mode: port-forwarding
    transportProtocol: quic
    # 因为公网 IP 不固定，所以这儿配置动态检测公网 IP
    publicAddress: autoHttpDetect
    httpDetector:
      url: https://ipinfo.io/ip
    publicPort: 60234
    listenPort: 60234

localExposedApps:
- appsocket: unix:/var/run/docker.sock
  appName: docker
```

启动 `wovenet`：

```
./wovenet run --config ./config.yaml
```

### docker-client 配置

创建 `config.yaml` 文件，内容如下：

```yaml
siteName: docker-client

crypto:
  # 这个 key 是用来加密敏感信息的，要求每个 wovenet 实例都要配置相同的值。
  # 特别注意：不要使用本示例提供的值，否则可能造成敏感信息泄露。
  key: "3iop#Mg6732ftg#4(ER"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
     # 特别注意：强烈建议修改 topic，相同 topic 的 wovenet 实例会共同组成一个 mesh 网络。
    topic: "kungze/wovenet/unix-to-tcp-67bhgerAB"

# 远端 docker app 监听在 unix socket 上，这儿转为了监听在 tcp socket 上了。
remoteApps:
- appName: docker
  siteName: docker-server
  localSocket: tcp:172.26.142.105:38902
```

启动 `wovenet`：

```
./wovenet run --config ./config.yaml
```

### 验证

在 docker-client 侧：

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
