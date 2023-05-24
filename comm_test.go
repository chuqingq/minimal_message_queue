package mmq

import (
	"io/ioutil"
	"log"
	"testing"
	"time"
)

var (
	caCert      []byte
	serverKey   []byte
	serverCert  []byte
	clientKey   []byte
	clientCert  []byte
	clientKey2  []byte
	clientCert2 []byte
	addr        string = "127.0.0.1:8080"
)

func initCert() {
	var err error
	// ca
	caCert, err = ioutil.ReadFile("certs/ca.cer")
	assert(err == nil)
	// server
	serverKey, err = ioutil.ReadFile("certs/server.key")
	assert(err == nil)
	serverCert, err = ioutil.ReadFile("certs/server.cer")
	assert(err == nil)
	// client
	clientKey, err = ioutil.ReadFile("certs/client.key")
	assert(err == nil)
	clientCert, err = ioutil.ReadFile("certs/client.cer")
	assert(err == nil)
	// invalid client
	clientKey2, err = ioutil.ReadFile("certs/another/client.key")
	assert(err == nil)
	clientCert2, err = ioutil.ReadFile("certs/another/client.cer")
	assert(err == nil)
}

func TestComm(t *testing.T) {
	initCert()
	// 初始化
	s := NewComm()
	c1 := NewComm()
	c2 := NewComm()
	c3 := NewComm()
	// 用例开始
	// startServer:server1
	log.Printf("====StartServer")
	s.StartServer(addr, serverKey, serverCert, caCert)
	// onStartServer
	msg := *s.Recv()
	log.Printf("s recv: %v", msg)
	assert(msg.String("cmd") == "onStartServer")
	assert(msg.Bool("success"))
	defer s.StopServer()

	// client1
	log.Printf("====startClient")
	c1.StartClient(addr, clientKey, clientCert, caCert)
	// onStartClient
	msg = *c1.Recv()
	log.Printf("c1 recv: %v", msg)
	assert(msg.String("cmd") == "onStartClient")
	assert(msg.Bool("success"))
	defer c1.StopClient()

	// client2
	log.Printf("====startClient2")
	c2.StartClient(addr, clientKey, clientCert, caCert)
	// defer c.StopClient("client2")
	// onStartClient
	msg = *c2.Recv()
	log.Printf("c2 recv: %v", msg)
	assert(msg.String("cmd") == "onStartClient")
	assert(msg.Bool("success"))
	// assert(msg["client"].(string) == "client2")

	// client3
	log.Printf("====startClient3 invalid")
	c3.StartClient(addr, clientKey2, clientCert2, caCert)
	// onStartClient
	msg = *c3.Recv()
	log.Printf("c3 recv: %v", msg)
	assert(msg.String("cmd") == "onStartClient")
	// assert(!msg.Bool("success"))

	// client1 subscribe
	log.Printf("====client1 subscribe client1Topic")
	c1.Subscribe("client1Topic")
	time.Sleep(100 * time.Millisecond) // 确保订阅成功后再发布消息

	// client2 publish
	log.Printf("====client2 send topic client1Topic")
	pmsg, err := MessageFromString(`{
		"data": {
			"key5": "value5"
		}
	}`)
	assert(err == nil)
	msg = *pmsg
	c2.Publish("client1Topic", &msg)

	// client1 recv
	msg = *c1.Recv()
	log.Printf("c1 recv msg: %v", msg)
	assert(msg.String("cmd") == "publish")
	assert(msg.String("topic") == "client1Topic")
	// stopClient:client2
	log.Printf("====stopClient client2")
	c2.StopClient()
}

func assert(b bool) {
	if !b {
		panic("assert failed!!!!")
	}
}
