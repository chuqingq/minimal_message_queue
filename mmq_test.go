package mmq

import (
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMmq(t *testing.T) {
	assert := assert.New(t)

	logger.SetLevel(logrus.DebugLevel)
	initCert(assert)

	// startServer
	log.Printf("====StartServer")
	s, err := NewServer(addr, serverKey, serverCert, caCert)
	assert.NoError(err)

	// onStartServer
	msg := *s.Recv()
	log.Printf("s recv: %v", msg)
	assert.Equal(msg.String("cmd"), "onStartServer")
	assert.True(msg.Bool("success"))
	defer s.Close()

	// client1
	log.Printf("====startClient")
	c1, err := NewClient(addr, clientKey, clientCert, caCert)
	assert.NoError(err)

	// onStartClient
	msg = *c1.Recv()
	log.Printf("c1 recv: %v", msg)
	assert.Equal(msg.String("cmd"), "onStartClient")
	assert.True(msg.Bool("success"))
	defer c1.Close()

	// msg = *s.Recv()
	// log.Printf("s recv: %v", msg)

	// client2
	log.Printf("====startClient2")
	c2, err := NewClient(addr, clientKey, clientCert, caCert)
	assert.NoError(err)
	// defer c.StopClient("client2")
	// onStartClient
	msg = *c2.Recv()
	log.Printf("c2 recv: %v", msg)
	assert.Equal(msg.String("cmd"), "onStartClient")
	assert.True(msg.Bool("success"))
	// assert(msg["client"].(string) == "client2")

	// client3
	log.Printf("====startClient3 invalid")
	c3, err := NewClient(addr, clientKey2, clientCert2, caCert)
	assert.NoError(err)
	// onStartClient
	msg = *c3.Recv()
	log.Printf("c3 recv: %v", msg)
	assert.Equal(msg.String("cmd"), "onStartClient")
	assert.True(msg.Bool("success"))

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
	assert.NoError(err)
	msg = *pmsg
	c2.Publish("client1Topic", &msg)

	// client1 recv
	msg = *c1.Recv()
	log.Printf("c1 recv msg: %v", msg)
	assert.Equal(msg.String("cmd"), "publish")
	assert.Equal(msg.String("topic"), "client1Topic")

	// stopClient:client2
	log.Printf("====stopClient client2")
	c2.Close()
}

func TestAssert(t *testing.T) {
	assert := assert.New(t)

	err := errors.New("this is an error")
	assert.NoError(err)

	assert.Equal(1+2, 4)
}

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

func initCert(assert *assert.Assertions) {
	var err error

	// ca
	caCert, err = os.ReadFile("certs/ca.cer")
	assert.NoError(err)

	// server
	serverKey, err = os.ReadFile("certs/server.key")
	assert.NoError(err)

	serverCert, err = os.ReadFile("certs/server.cer")
	assert.NoError(err)

	// client
	clientKey, err = os.ReadFile("certs/client.key")
	assert.NoError(err)

	clientCert, err = os.ReadFile("certs/client.cer")
	assert.NoError(err)

	// invalid client
	clientKey2, err = os.ReadFile("certs/another/client.key")
	assert.NoError(err)

	clientCert2, err = os.ReadFile("certs/another/client.cer")
	assert.NoError(err)
}
