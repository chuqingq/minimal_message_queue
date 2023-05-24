package mmq

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"net"
	"strings"
	"time"
)

// mmqClient 客户端
type mmqClient struct {
	conn    net.Conn
	dec     *json.Decoder
	enc     *json.Encoder
	topics  *mmqTopics
	server  *mmqServer // server==nil表示是peer
	outChan chan Message
}

// NewClient 启动客户端
func NewClient(addr string, key, cert, ca []byte) (*mmqClient, error) {
	logger.Debugf("NewClient(addr: %v)", addr)
	client, err := connect(addr, key, cert, ca)
	if err != nil {
		errMsg := fmt.Sprintf("StartClient error: %v", err)
		logger.Error(errMsg)
		client.onStartClient(false, errMsg)
		return nil, err
	}
	logger.Debugf("StartClient[%p] success", client)
	client.onStartClient(true, "")
	return client, nil
}

// onStartClient 返回启动客户端事件
func (c *mmqClient) onStartClient(success bool, msg string) {
	m := NewMessage()
	m.Set("cmd", "onStartClient")
	m.Set("success", success)
	m.Set("msg", msg)
	sendMessage(c.outChan, m)
}

func (c *mmqClient) IsClientAlive() bool {
	return c.conn != nil
}

// Subscribe 客户端订阅主题
func (c *mmqClient) Subscribe(topics string) {
	logger.Debugf("client[%p].Subscribe(%v)", c, topics)
	m := NewMessage()
	m.Set("cmd", "subscribe")
	m.Set("topics", topics)
	if c != nil && c.enc != nil {
		c.send(m)
	} else {
		logger.Errorf("client[%p] disconnected", c)
	}
}

// Unsubscribe 客户端取消订阅主题
func (c *mmqClient) Unsubscribe(topics string) {
	logger.Debugf("client[%p].Unsubscribe(%v)", c, topics)
	m := NewMessage()
	m.Set("cmd", "unsubscribe")
	m.Set("topics", topics)
	if c != nil && c.enc != nil {
		c.send(m)
	} else {
		logger.Errorf("client[%p] disconnected", c)
	}
}

// Publish 客户端发布消息
func (c *mmqClient) Publish(topic string, m *Message) {
	logger.Debugf("client[%p].Publish(%v, %v)", c, topic, m)
	m.Set("cmd", "publish")
	m.Set("topic", topic)
	if c != nil && c.enc != nil {
		c.send(m)
	} else {
		logger.Errorf("client[%p] disconnected", c)
	}
}

func (c *mmqClient) Recv() *Message {
	return recvWithTimeout(c.outChan, -1)
}

func (c *mmqClient) TryRecv() *Message {
	return recvWithTimeout(c.outChan, 0)
}

func (c *mmqClient) Close() {
	logger.Debugf("client[%p] close", c)
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	if c.server != nil {
		c.server.delPeer(c)
	}
}

func setKeepAlive(conn net.Conn, d time.Duration) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(d)
}

func connect(addr string, key, cert, ca []byte) (*mmqClient, error) {
	// ca
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	// clientCert
	cliCrt, err := tls.X509KeyPair(cert, key)
	if err != nil {
		logger.Debugf("X509KeyPair() error: %v", err)
		return nil, err
	}
	// tlsConfig
	tlsConfig := &tls.Config{
		RootCAs:            pool,
		Certificates:       []tls.Certificate{cliCrt},
		InsecureSkipVerify: true,
	}
	// dial
	tcpConn, err := net.Dial("tcp", addr)
	if err != nil {
		logger.Debugf("Dial(%v) error: %v", addr, err)
		return nil, err
	}
	// setkeepalive
	setKeepAlive(tcpConn, 10*time.Second)
	// setKeepAlive(tcpConn.(*net.TCPConn), 10, 3, 10)
	// tlsConn
	conn := tls.Client(tcpConn, tlsConfig)
	// handshake
	err = conn.Handshake()
	if err != nil {
		logger.Debugf("connect handshake error: %v", err)
		return nil, err
	}
	logger.Debugf("connect handshake success")
	client := &mmqClient{
		conn: conn,
		topics: &mmqTopics{
			topics: make(map[string]interface{}, 16),
		},
		outChan: make(chan Message, 128),
	}
	client.start()
	return client, nil
}

// 建连完成后，启动交互
func (c *mmqClient) start() {
	c.enc = json.NewEncoder(c.conn)
	go func() {
		c.dec = json.NewDecoder(c.conn)
		var err error
		for {
			msg := &Message{}
			err = c.dec.Decode(msg) // simplejson.Json实现了Unmarshaler interface，因此可以正确解析
			if err != nil {
				// 需要判断错误类型。如果是本端断开，上报onServerDisconnected；如果是远端断开，上报onClientDisconnected
				if c.server == nil {
					logger.Debugf("client[%p] disconnected: %v", c, err)
				} else {
					logger.Debugf("peer[%p] disconnected: %v", c, err)
				}
				errStr := err.Error()
				if strings.Contains(errStr, "use of closed network connection") {
					// c.close()
				} else /* if strings.Contains(errStr, "EOF")*/ {
					if c.server == nil {
						// 如果是客户端，需要通知业务连接已断开
						c.onStartClient(false, errStr)
					}
					//  else {
					// 	// 如果是服务端，无法重连，直接断开
					// }

				}
				c.Close()
				return
			}
			c.handleCmd(msg)
		}
	}()
}

func (c *mmqClient) handleCmd(msg *Message) {
	logger.Debugf("client/peer[%p].handleCmd(%v)", c, msg)
	cmd := msg.Get("cmd").MustString()
	switch cmd {
	case "subscribe":
		// 只可能是peer
		c.topics.add(msg.Get("topics").MustString())
	case "unsubscribe":
		// 只可能是peer
		c.topics.remove(msg.Get("topics").MustString())
	case "publish":
		// 1. 如果是client，则是server端根据订阅分发过来的，直接发给业务
		if c.server == nil {
			sendMessage(c.outChan, msg)
			break
		}
		// 2. 如果是peer，需要对应的server分发
		topic := msg.Get("topic").MustString()
		c.server.dispatchTopic(topic, msg)
	default:
		logger.Debugf("unknown cmd: %v", cmd)
	}
}

func (c *mmqClient) send(v interface{}) error {
	logger.Debugf("client/peer[%p].Send() %v", c, v)
	return c.enc.Encode(v)
}
