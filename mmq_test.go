package mmq

import (
	"os"
	"testing"

	sjson "github.com/chuqingq/simple-json"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMmq(t *testing.T) {
	assert := assert.New(t)

	logger.SetLevel(logrus.DebugLevel)
	logger.SetReportCaller(true)
	initCert(assert)

	checkChan := make(chan sjson.Json, 4)

	// startServer
	logger.Debugf("====StartServer")
	s := NewServer(addr)
	s.SetTLS(serverKey, serverCert, caCert)
	err := s.Start()
	assert.NoError(err)
	defer s.Stop()

	// client1
	logger.Debugf("====startClient")
	c1 := NewClient(addr)
	c1.SetTLS(clientKey, clientCert, caCert)
	c1.SetOnMsgRecv(func(c *Client, topic string, msg *sjson.Json, err error) {
		logger.Debugf("c1 recv msg: %v", msg)
		checkChan <- *msg
	})
	err = c1.Start()
	assert.NoError(err)
	defer c1.Stop()

	// client2
	logger.Debugf("====startClient2")
	c2 := NewClient(addr)
	c2.SetTLS(clientKey, clientCert, caCert)
	c2.SetOnMsgRecv(func(c *Client, topic string, msg *sjson.Json, err error) {
		logger.Fatalf("c2 recv msg: %v", msg)
		// checkChan <- msg
	})
	err = c2.Start()
	assert.NoError(err)
	defer c2.Stop()

	// client1 subscribe
	logger.Debugf("====client1 subscribe client1Topic")
	c1.Subscribe([]string{"client1Topic"})

	// client2 publish
	logger.Debugf("====client2 send topic client1Topic")
	pmsg, err := sjson.FromString(`{
		"data": {
			"key5": "value5"
		}
	}`)
	assert.NoError(err)
	msg := *pmsg
	c2.Publish("client1Topic", &msg)

	// client1 recv
	msg = <-checkChan
	assert.Equal(msg.Get("data.key5").MustString(), "value5")
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
