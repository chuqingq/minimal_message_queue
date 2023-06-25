package mmq

import (
	sjson "github.com/chuqingq/simple-json"
	"github.com/chuqingq/simple-tcpjson"
)

// Client 客户端
type Client struct {
	client          *tcpjson.Client
	OnClientMsgRecv OnClientMsgRecv
}

type OnClientMsgRecv func(c *Client, topic string, msg *sjson.Json, err error)

// NewClient 启动客户端
func NewClient(addr string) *Client {
	return &Client{
		client: tcpjson.NewClient(addr),
	}
}

// SetTLS 设置TLS
func (c *Client) SetTLS(key, cert, ca []byte) *Client {
	c.client.SetTLS(key, cert, ca)
	return c
}

func (c *Client) SetOnMsgRecv(handler OnClientMsgRecv) *Client {
	c.client.SetOnMsgRecv(func(client *tcpjson.Client, msg *sjson.Json, err error) {
		if err != nil || msg == nil {
			return
		}
		topic := msg.Get("topic").MustString()
		m := msg.Get("msg")
		handler(c, topic, m, err)
	})
	return c
}

func (c *Client) Start() error {
	return c.client.Start()
}

// Subscribe 客户端订阅主题
func (c *Client) Subscribe(topics []string) error {
	logger.Debugf("client[%p].Subscribe(%v)", c, topics)
	m := &sjson.Json{}
	m.Set("cmd", "subscribe")
	m.Set("topics", topics)
	return c.client.Send(m)
}

// Unsubscribe 客户端取消订阅主题
func (c *Client) Unsubscribe(topics []string) error {
	logger.Debugf("client[%p].Unsubscribe(%v)", c, topics)
	m := &sjson.Json{}
	m.Set("cmd", "unsubscribe")
	m.Set("topics", topics)
	return c.client.Send(m)
}

// Publish 客户端发布消息
func (c *Client) Publish(topic string, m *sjson.Json) error {
	logger.Debugf("client[%p].Publish(%v, %v)", c, topic, m)
	msg := &sjson.Json{}
	msg.Set("cmd", "publish")
	msg.Set("topic", topic)
	msg.Set("msg", m)
	return c.client.Send(msg)
}

func (c *Client) Stop() {
	logger.Debugf("client[%p].Stop", c)
	c.client.Stop()
}
