### 通信模块接口

1. 所有消息是异步的，即不需要应答。
2. 业务或通信模块根据自己的需要发送应答，比如startServer对应startServerResult。
3. 业务可以同时作为server或client，数量不限。
4. cmd为data，表示是业务给业务发送的数据；否则，表示业务和comm之间发送的事件

#### 业务作为server

##### 命令：启动server 123
方向：业务->comm

```json
{
    "cmd": "startServer",
    "server": "123",
    "addr": "127.0.0.1:8080",
    "key": "",
    "cert": "",
    "ca": ""
}
```

##### 命令：server启动结果
方向：comm->业务

```json
{
    "cmd": "onStartServer",
    "server": "123",
    "success": true,
    "msg": ""
}
```

##### 命令：停止server 123
方向：业务->comm

```json
{
    "cmd": "stopServer",
    "server": "123"
}
```

##### 命令：server123收到新client456的连接
方向：comm->业务

```json
{
    "cmd": "onClientConnected",
    "server": "123",
    "client": "456"
}
```

##### 命令：server123收到client456断连
方向：comm->业务

```json
{
    "cmd": "onClientDisconnected",
    "server": "123",
    "client": "456"
}
```

#### 业务作为client

##### 命令：启动client456
方向：业务->comm

```json
{
    "cmd": "startClient",
    "addr": "127.0.0.1:8080",
    "client": "456",
    "key": "",
    "cert": "",
    "ca": ""
}
```

##### 命令：启动client456结果
方向：comm->业务

```json
{
    "cmd": "onStartClient",
    "client": "456",
    "success": true,
    "msg": ""
}
```

##### 命令：停止client连接
方向：业务->comm

```json
{
    "cmd": "stopClient",
    "client": "456"
}
```

##### 命令：server断开连接
方向：comm->业务

```json
{
    "cmd": "onServerDisconnected",
    // "server": "123",
    "client": "456"
}
```

#### 订阅

##### 命令：client向server订阅topic

说明：业务订阅cmd:data & topic:someTopicName的消息。comm收到符合条件的消息会转发给对应的业务。
方向：业务->comm

```json
{
    "cmd": "subscribe",
    "topic": "someTopicName",
    "client": "456"
}
```

##### 命令：client向server取消订阅topic

说明：业务取消订阅。
方向：业务->comm

```json
{
    "cmd": "unsubscribe",
    "topic": "someTopicName",
    "client": "456"
}
```

#### 收发消息

##### 命令：server给client发消息
方向：业务123->comm；comm->业务456

```json
{
    "cmd": "data",
    "from": "123",
    "to": "456",
    "topic": "someTopicName",
    "data": {
    }
}
```

#### 其他

##### 命令：查询连接是否在线
方向：业务->comm

```json
{
    "cmd": "isAlive",
    "client": "123",
}
```

#### 命令：返回连接是否在线查询结果
方向：comm->业务

```json
{
    "cmd": "onIsAlive",
    "client": "123",
    "isAlive": true
}
```

