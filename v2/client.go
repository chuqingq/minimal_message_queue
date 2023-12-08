package mmq

import (
	"strings"

	"github.com/chuqingq/minimal_message_queue/v2/tcp"
)

// Client 客户端
type Client struct {
	client          *tcp.Client
	OnClientMsgRecv OnClientMsgRecv
}

type OnClientMsgRecv func(c *Client, topic string, data []byte, err error)

// NewClient 启动客户端
func NewClient(addr string) *Client {
	return &Client{
		client: tcp.NewClient(addr),
	}
}

func (c *Client) SetOnMsgRecv(handler OnClientMsgRecv) *Client {
	c.client.SetOnMsgRecv(func(client *tcp.Client, data []byte, err error) {
		if err != nil || data == nil {
			return
		}
		cmd := &Command{}
		cmd.FromBytes(data)
		handler(c, cmd.Topic, cmd.Data, err)
	})
	return c
}

func (c *Client) Start() error {
	return c.client.Start()
}

// Subscribe 客户端订阅主题
func (c *Client) Subscribe(topics []string) error {
	logger.Debugf("client[%p].Subscribe(%v)", c, topics)
	return c.sendCommand(&Command{
		Cmd:   CmdSubscribe,
		Topic: strings.Join(topics, ","),
	})
}

// Unsubscribe 客户端取消订阅主题
func (c *Client) Unsubscribe(topics []string) error {
	logger.Debugf("client[%p].Unsubscribe(%v)", c, topics)
	return c.sendCommand(&Command{
		Cmd:   CmdUnsubscribe,
		Topic: strings.Join(topics, ","),
	})
}

// Publish 客户端发布消息
func (c *Client) Publish(topic string, data []byte) error {
	logger.Debugf("client[%p].Publish(%v, %v)", c, topic, string(data))
	return c.sendCommand(&Command{
		Cmd:   CmdPublish,
		Topic: topic,
		Data:  data,
	})
}

func (c *Client) sendCommand(cmd *Command) error {
	b, err := cmd.ToBytes()
	if err != nil {
		return err
	}
	return c.client.Send(b)
}

func (c *Client) Stop() {
	logger.Debugf("client[%p].Stop", c)
	c.client.Stop()
}
