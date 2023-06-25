# minimal_message_queue

A simple message queue module in Go, based on tcp/json.

## Features


## TODO

## go doc

```
package mmq // import "mmq"


TYPES

type Client struct {
	OnClientMsgRecv OnClientMsgRecv
	// Has unexported fields.
}
    Client 客户端

func NewClient(addr string) *Client
    NewClient 启动客户端

func (c *Client) Publish(topic string, m *sjson.Json) error
    Publish 客户端发布消息

func (c *Client) SetOnMsgRecv(handler OnClientMsgRecv) *Client

func (c *Client) SetTLS(key, cert, ca []byte) *Client
    SetTLS 设置TLS

func (c *Client) Start() error

func (c *Client) Stop()

func (c *Client) Subscribe(topics []string) error
    Subscribe 客户端订阅主题

func (c *Client) Unsubscribe(topics []string) error
    Unsubscribe 客户端取消订阅主题

type MatchTopicFunc func(pubtopic, subtopic string) bool

type OnClientMsgRecv func(c *Client, topic string, msg *sjson.Json, err error)

type Server struct {
	MatchTopicFunc MatchTopicFunc

	// Has unexported fields.
}
    Server 服务端

func NewServer(addr string) *Server
    NewServer 创建并启动服务端

func (s *Server) SetCluster(addr string) *Server

func (s *Server) SetMatchTopicFunc(match MatchTopicFunc) *Server

func (s *Server) SetTLS(key, cert, ca []byte) *Server

func (s *Server) Start() error

func (s *Server) Stop()
```
