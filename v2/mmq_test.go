package mmq

import (
	"log"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// TODO tcp server需要实现keepalive

func TestMmq(t *testing.T) {
	assert := assert.New(t)

	logger.SetLevel(logrus.DebugLevel)
	logger.SetReportCaller(true)
	// initCert(assert)

	checkChan := make(chan []byte, 4)

	// startServer
	log.Printf("====StartServer")
	s := NewServer(addr)
	// s.SetTLS(serverKey, serverCert, caCert)
	err := s.Start()
	assert.NoError(err)
	defer s.Stop()

	// client1
	log.Printf("====startClient")
	c1 := NewClient(addr)
	// c1.SetTLS(clientKey, clientCert, caCert)
	c1.SetOnMsgRecv(func(c *Client, topic string, msg []byte, err error) {
		log.Printf("c1 recv msg: %v", string(msg))
		checkChan <- msg
	})
	err = c1.Start()
	assert.NoError(err)
	defer c1.Stop()

	// client2
	log.Printf("====startClient2")
	c2 := NewClient(addr)
	// c2.SetTLS(clientKey, clientCert, caCert)
	c2.SetOnMsgRecv(func(c *Client, topic string, msg []byte, err error) {
		log.Fatalf("c2 recv msg: %v", string(msg))
		// checkChan <- msg
	})
	err = c2.Start()
	assert.NoError(err)
	defer c2.Stop()

	// client1 subscribe
	log.Printf("====client1 subscribe client1Topic")
	c1.Subscribe([]string{"client1Topic"})

	time.Sleep(20 * time.Millisecond)

	// client2 publish
	log.Printf("====client2 send topic client1Topic")
	msg := []byte("value5")
	c2.Publish("client1Topic", msg)

	// client1 recv
	msg2 := <-checkChan
	assert.Equal(msg, msg2)
}

var (
	// 	caCert      []byte
	// 	serverKey   []byte
	// 	serverCert  []byte
	// 	clientKey   []byte
	// 	clientCert  []byte
	// 	clientKey2  []byte
	// 	clientCert2 []byte
	addr string = "127.0.0.1:8080"
)

// func initCert(assert *assert.Assertions) {
// 	var err error

// 	// ca
// 	caCert, err = os.ReadFile("certs/ca.cer")
// 	assert.NoError(err)

// 	// server
// 	serverKey, err = os.ReadFile("certs/server.key")
// 	assert.NoError(err)

// 	serverCert, err = os.ReadFile("certs/server.cer")
// 	assert.NoError(err)

// 	// client
// 	clientKey, err = os.ReadFile("certs/client.key")
// 	assert.NoError(err)

// 	clientCert, err = os.ReadFile("certs/client.cer")
// 	assert.NoError(err)

// 	// invalid client
// 	clientKey2, err = os.ReadFile("certs/another/client.key")
// 	assert.NoError(err)

// 	clientCert2, err = os.ReadFile("certs/another/client.cer")
// 	assert.NoError(err)
// }
