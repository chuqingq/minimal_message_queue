### TODO

- [ ] 无法解决：linux交叉编译windows的comm dll报错：gcc: error: unrecognized command-line option ‘-mthreads’; did you mean ‘-pthread’? 原因是cgo的交叉编译无法在linux编译windows的动态库。交叉编译可执行程序是可以的
- [x] 客户端和服务端的网络API
- [x] 可以多平台使用
- [x] 封装给java使用。目前是跨进程，宿主可以是java、python、c++等任何语言，通过标准输入输出来操作cgi可执行文件
- [x] java需要进行解包：使用gson或fastjson的jsonreader
- [x] 小余：windows的问题
- [x] 小余：1.tls加密功能，让key、cert、ca生效。2.需要几套、如何生成key、cert、ca，输出文档
- [ ] comm客户端连接后，先发个消息给服务端，一方面说明自己是谁，另一方面确认TLS handshake 成功
- [ ] comm和云侧的websocket之间怎么通信？
- [ ] 验证非ca签名的证书，无法建立通信
- [ ] 如果有相同的clientid连接到server，会如何？
- [ ] 把main.go中的客户端逻辑和服务端逻辑分开
- [ ] 通信模块改成消息总线，简化架构。比如nsq这种。结论：不用消息总线，每个节点上之保留一个中心节点即可，比如应用和设备网关都连接基线
- [ ] 加解密：最好直接用TLS。tls+io.reader+json。TLS需要双向认证。
- [ ] 异常断连后重连
- [ ] 心跳
- [ ] server.listen后等accept协程启动后再给业务回serverOnline
- [ ] cgi.Stop确保所有server都停止
- [ ] comm.go中的startServer等不带json的接口可以直接暴露到so中；收发数据的Data接口因为用到json，需要在so中封装一个接口暴露json字符串
- [ ] 生成证书服务

- [ ] Comm名字改为mq
- [ ] StartClient和StartServer分开
- [ ] 去掉RegisterMessageCallback，因为回调的方式和Recv的方式冲突，可能导致Recv没收到消息。回调完全可以让业务自己起协程去Recv再callback
- [ ] 考虑使用对称密钥代替TLS。简化
- [ ] 考虑使用mDNS代替tcp服务端。
