# 借助 wovenet 应对存在大量动态变化的端口的系统（如：VDI 系统）

[English](./README.md)

有些系统，在对外提供服务时需要暴露多个端口，而且这些端口可能是动态变化的。如：VDI(Virtual Desktop Infrastructure)平台。

VDI 也被称为云桌面系统，在 VDI 系统每个云桌面就是系统内一个虚机，这些虚机通过 RDP 服务或者 SPICE 服务供远端用户使用。每个虚机都有自己的 RDP 服务，或者 SPICE 服务。因此一套 VDI 系统要想在公网上被客户远程使用，就需要在公网上暴露大量端口，而且考虑到云桌面的创建，删除，开机，关机，这些端口还是动态变化的，这些云桌面的 RDP/SPICE 端口如何和公网端口映射，也是一个巨大的挑战。

## 借助 wovenet 减少端口暴露数量，无惧端口变化

由于我在写这篇文档时并没有可用的 VDI 系统，而且为了方便读者复现该示例结果，我使用 `nc` 命令启动多个服务监听多个端口模拟多个 VDI 虚机。

### 环境信息

#### 数据中心服务侧

公网 IP：115.190.109.199

nc 服务：

* `nc` 服务1：10.0.2.128:5900
* `nc` 服务2：10.0.2.129:5900
* `nc` 服务3：10.0.2.129:5926

#### 客户端侧

三个客户端，分别使用：client1，client2，client3 进行标识。

### 数据中心服务侧配置

创建 `config.yaml` 文件，内容如下：

```yaml
siteName: data-center

crypto:
  # 这个 key 是用来加密敏感信息的，要求每个 wovenet 实例都要配置相同的值。
  # 特别注意：不要使用本示例提供的值，否则可能造成敏感信息泄露。
  key: "YHU8D(!dse4UIdewrd56D"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # 特别注意：强烈建议修改 topic，相同 topic 的 wovenet 实例会共同组成一个 mesh 网络。
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

**上面配置，在数据中心服务侧只需要暴露一个 UDP 25890 端口。**

###### 启动 `wovenet`：

```
./wovenet run --config ./config.yaml
```

###### 启动 nc 服务：

* 在 10.0.2.128 上：

```
nc -l 5900
```

* 在 10.0.2.129 上启动两个 nc 服务：

nc 服务1：

```
nc -l 5900
```

nc 服务2：

```
nc -l 5926
```

### 客户端测配置

三个客户端分别创建 `config.yaml` 文件，内容如下：

```yaml
# 其他两个客户端修改 siteName，分别为 client2，client3
siteName: client1

crypto:
  # 这个 key 是用来加密敏感信息的，要求每个 wovenet 实例都要配置相同的值。
  # 特别注意：不要使用本示例提供的值，否则可能造成敏感信息泄露。
  key: "YHU8D(!dse4UIdewrd56D"

logger:
  level: DEBUG
  file: ""
  format: json

messageChannel:
  protocol: mqtt
  mqtt:
    brokerServer: mqtt://mqtt.eclipseprojects.io:1883
    # 特别注意：强烈建议修改 topic，相同 topic 的 wovenet 实例会共同组成一个 mesh 网络。
    topic: "kungze/wovenet/multiple-port-nywgf67s345od"

remoteApps:
- siteName: data-center
  appName: nc
  # 其他两个客户端需要修改 appSocket，分别为：tcp:10.0.2.129:5900，tcp:10.0.2.129:5926
  appSocket: tcp:10.0.2.128:5900
  localSocket: tcp:127.0.0.1:15900
```

###### 启动 `wovenet`：

三个客户端分别执行下面命令

```
./wovenet run --config ./config.yaml
```

### 验证

分别在 client1，client2，client3 终端输入以下内容，观察数据中心侧输出：

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

* client3

```
$ nc 127.0.0.1 15900
hello, I am client3.
```

在数据中心侧，对应的 nc 服务会有相对应的输出：

* 10.0.2.128 上 nc 输出：

```
$ nc -l 5900
hello, I am client1.
```

* 10.0.2.129 上第一个 nc 服务输出：

```
$ nc -l 5900
hello, I am client2.
```

* 10.0.2.129 上第二个 nc 服务输出：

```
$ nc -l 5926
hello，I am client3.
```
