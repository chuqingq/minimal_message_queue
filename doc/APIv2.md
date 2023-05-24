# 消息总线模块（通信模块）接口

1. 具备消息总线广播模式，即A发布topic消息，所有订阅topic的客户端都能够收到此消息。
2. 消息不缓存，如果某消息发布时订阅者不存在，则即使后续订阅也不会再收到此消息。
3. 不是所有信令都有响应，详细参考如下API文档。例如startServer的响应是onStartServer，stopServer无响应、默认成功。
4. 订阅topic支持模糊匹配，如下两种方式
    1. \*通配任意一个word，例如：发布的"a.b.c"能匹配上订阅的"a.b.*"
    2. 尾部#通配任意多个word，例如：发布的"a.b.c"能匹配上订阅的"a.#"

## commServer：启停服务

### 命令：启动服务
方向：业务->comm

```json
{
    "cmd": "startServer",
    "addr": "127.0.0.1:8080",
    "key": "",
    "cert": "",
    "ca": ""
}
```

### 命令：启动结果
方向：comm->业务

```json
{
    "cmd": "onStartServer",
    "success": true,
    "msg": ""
}
```

### 命令：停止服务
方向：业务->comm

```json
{
    "cmd": "stopServer",
}
```

## commClient：连接服务

### 命令：连接服务

方向：业务->comm->commServer
备注：重连由消息总线负责；重连后通讯库会保留之前已订阅的topics。如果是客户端重新connectServer，会清除已订阅的topics，需要自行重新订阅

```json
{
    "cmd": "startClient",
    "addr": "127.0.0.1:8080",
    "key": "",
    "cert": "",
    "ca": ""
}
```

### 命令：连接服务结果

方向：comm->业务
备注：通讯库不停重连，无论成功或失败都通知业务，业务根据需要是否重新订阅。

```json
{
    "cmd": "onStartClient",
    "success": true,
    "msg": ""
}
```

### 命令：断开服务
方向：业务->comm

```json
{
    "cmd": "stopClient"
}
```

## 订阅和退订

订阅和发布是commClient之间通信的唯一方式。

### 命令：订阅

说明：业务订阅topics的消息。comm收到符合条件的消息会转发给对应的业务。
方向：业务->comm->commServer
备注：\*通配任意一个word，例如发布的"a.b.c"能匹配上订阅的"a.b.*"；尾部#，通配任意多个word，例如发布的"a.b.c"能匹配上订阅的"a.#"

```json
{
    "cmd": "subscribe",
    "topics": "someTopicName1,someTopicName2,someTopicName3"
}
```

### 命令：取消订阅

说明：业务取消订阅。
方向：业务->comm->commServer

```json
{
    "cmd": "unsubscribe",
    "topics": "someTopicName1,someTopicName2,someTopicName3"
}
```

## 发布

### 命令：发布和接收消息

说明：业务向消息总线服务发布topic消息
方向：业务->comm->commServer->comm->业务2

```json
{
    "cmd": "publish",
    "topic": "someTopicName1",
    "data": {
        "key1": "value1",
        "key2": "value2"
    }
}
```

## commCloud：云模块

commCloud云模块作为commServer的独立功能模块。
在commServer侧使用连接、断连等接口；
在commClient侧使用云请求等接口

### 连接云（作为服务端）

说明：业务请求消息总线通过websocket连接云。
方向：业务->comm(commServer)

```json
{
    "cmd": "connectCloud",
    "cloudClientID": "cloudClientID1", // 给云侧标识本地是谁
    "url": "wss://junhe-iot.com/websocket",
    "key": "",
    "cert": "",
    "ca": ""
}
```

### 连接云结果

说明：消息总线通知业务连接云结果。
方向：comm->业务

```json
{
    "cmd": "onConnectCloud",
    "success": true,
    "msg": ""
}
```

### 断开云

说明：业务请求消息总线断开云。
方向：业务->comm

```json
{
    "cmd": "disconnectCloud"
}
```

### 请求云（作为客户端）

说明：业务向云请求消息。
方向：业务->comm->commServer

```json
{
    "cmd": "publish",
    "topic": "cloudRequest",
    "data": {
        "reqid": "123",
        "key1": "value1",
        "key2": "value2"
    }
}
```

### 云响应

说明：消息总线从云侧接收到的消息直接发布到消息总线上。因此需要云侧设置cmd为publish、topic为指定的主题名。
方向：cloud->commServer->comm->业务
