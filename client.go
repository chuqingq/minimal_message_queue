package mmq

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"

	"mmq/log"

	// "junheiot/log"
	"net"
	"strings"
	"time"
	// "junheiot/util"
)

// StartClient comm启动客户端
func (comm *Comm) StartClient(addr string, key, cert, ca []byte) error {
	log.Logger().Debugf("comm[%p].StartClient(addr: %v)", comm, addr)
	client, err := connect(comm, addr, key, cert, ca)
	if err != nil {
		errMsg := fmt.Sprintf("StartClient error: %v", err)
		log.Logger().Error(errMsg)
		comm.onStartClient(false, errMsg)
		return err
	}
	comm.client = client
	log.Logger().Debugf("StartClient success")
	comm.onStartClient(true, "")
	return nil
}

// onStartClient 返回启动客户端事件
func (comm *Comm) onStartClient(success bool, msg string) {
	m := NewMessage()
	m.Set("cmd", "onStartClient")
	m.Set("success", success)
	m.Set("msg", msg)
	comm.input(m)
}

// StopClient 停止comm客户端
func (comm *Comm) StopClient() {
	log.Logger().Debugf("comm[%p].StopClient()", comm)
	if comm.client != nil {
		comm.client.close()
	}
}

func (comm *Comm) IsClientAlive() bool {
	return comm.client != nil
}

// Subscribe comm客户端订阅主题
func (comm *Comm) Subscribe(topics string) {
	log.Logger().Debugf("comm[%p]client.Subscribe(%v)", comm, topics)
	m := NewMessage()
	m.Set("cmd", "subscribe")
	m.Set("topics", topics)
	if comm.client != nil && comm.client.enc != nil {
		comm.client.send(m)
	} else {
		log.Logger().Errorf("comm is disconnected")
	}
}

// Unsubscribe comm客户端取消订阅主题
func (comm *Comm) Unsubscribe(topics string) {
	log.Logger().Debugf("comm[%p]client.Unsubscribe(%v)", comm, topics)
	m := NewMessage()
	m.Set("cmd", "unsubscribe")
	m.Set("topics", topics)
	if comm.client != nil && comm.client.enc != nil {
		comm.client.send(m)
	} else {
		log.Logger().Errorf("comm is disconnected")
	}
}

// Publish comm客户端发布消息
func (comm *Comm) Publish(topic string, m *Message) {
	log.Logger().Debugf("comm[%p]client.Publish(%v, %v)", comm, topic, m)
	m.Set("cmd", "publish")
	m.Set("topic", topic)
	if comm.client != nil && comm.client.enc != nil {
		comm.client.send(m)
	} else {
		log.Logger().Errorf("comm is disconnected")
	}
}

// commClient comm客户端
type commClient struct {
	conn   net.Conn
	dec    *json.Decoder
	enc    *json.Encoder
	topics *commTopics
	server *commServer // server==nil表示是peer
	comm   *Comm
}

func setKeepAlive(conn net.Conn, d time.Duration) {
	tcpConn := conn.(*net.TCPConn)
	tcpConn.SetKeepAlive(true)
	tcpConn.SetKeepAlivePeriod(d)
}

func connect(comm *Comm, addr string, key, cert, ca []byte) (*commClient, error) {
	// ca
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	// clientCert
	cliCrt, err := tls.X509KeyPair(cert, key)
	if err != nil {
		log.Logger().Debugf("X509KeyPair() error: %v", err)
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
		log.Logger().Debugf("Dial(%v) error: %v", addr, err)
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
		log.Logger().Debugf("connect handshake error: %v", err)
		return nil, err
	}
	log.Logger().Debugf("connect handshake success")
	client := &commClient{
		conn: conn,
		topics: &commTopics{
			topics: make(map[string]interface{}, 16),
		},
	}
	client.start(comm)
	return client, nil
}

func (c *commClient) start(comm *Comm) {
	c.comm = comm
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
					log.Logger().Debugf("comm[%p]client disconnected: %v", comm, err)
				} else {
					log.Logger().Debugf("comm[%p]peer disconnected: %v", comm, err)
				}
				errStr := err.Error()
				if strings.Contains(errStr, "use of closed network connection") {
					// c.close()
				} else /* if strings.Contains(errStr, "EOF")*/ {
					if c.server == nil {
						// 如果是客户端，需要通知业务连接已断开
						c.comm.onStartClient(false, errStr)
					} else {
						// 如果是服务端，无法重连，直接断开
					}

				}
				c.close()
				return
			}
			c.handleCmd(msg)
		}
	}()
}

func (c *commClient) handleCmd(msg *Message) {
	log.Logger().Debugf("comm[%p]client/peer.handleCmd(%v)", c.comm, msg)
	cmd := msg.Get("cmd").MustString()
	switch cmd {
	case "subscribe":
		c.topics.add( /*(*msg)["topics"].(string)*/ msg.Get("topics").MustString())
	case "unsubscribe":
		c.topics.remove( /*(*msg)["topics"].(string)*/ msg.Get("topics").MustString())
	case "publish":
		// 0. cmd=publish && topic = cloudRequest 转发给云
		// if /*(*msg)["topic"].(string)*/msg.Get("topic").MustString() == "cloudRequest" {
		// 	c.comm.sendToCloud(msg)
		// 	break
		// }
		// 1. 本端是客户端，则是server端根据订阅分发过来的，直接发给业务
		if c.server == nil {
			c.comm.input(msg)
			break
		}
		// 2. 本端是服务端，需要分发
		topic := /*(*msg)["topic"].(string)*/ msg.Get("topic").MustString()
		c.comm.dispatchTopic(topic, msg)
	default:
		log.Logger().Debugf("unknown cmd: %v", cmd)
	}
}

func (c *commClient) send(v interface{}) error {
	log.Logger().Debugf("comm[%p]client/peer.Send() %v", c.comm, v)
	return c.enc.Encode(v)
}

func (c *commClient) close() {
	c.conn.Close()
}
