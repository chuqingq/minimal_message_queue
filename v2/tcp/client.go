package tcp

import (
	"encoding/gob"
	"log"
	"net"
)

type Client struct {
	ServerAddr string

	OnStateChange OnClientStateChange
	OnMsgRecv     OnClientMsgRecv

	Conn    net.Conn
	encoder *gob.Encoder
	decoder *gob.Decoder

	State ClientState
}

func NewClient(serveraddr string) *Client {
	return &Client{ServerAddr: serveraddr, State: ClientDisconnected}
}

func (c *Client) SetOnStateChange(handler OnClientStateChange) *Client {
	c.OnStateChange = handler
	return c
}

func (c *Client) SetOnMsgRecv(handler OnClientMsgRecv) *Client {
	c.OnMsgRecv = handler
	return c
}

func (c *Client) Start() error {
	var err error
	c.Conn, err = net.Dial("tcp", c.ServerAddr)
	if err != nil {
		return err
	}
	c.State = ClientConnected

	c.encoder = gob.NewEncoder(c.Conn)
	c.decoder = gob.NewDecoder(c.Conn)
	go c.loop()
	return nil
}

func (c *Client) loop() {
	msg := make([]byte, 0, 102400)
	for {
		err := c.decoder.Decode(&msg)
		if err != nil {
			log.Printf("tcp client recv error: %v", err)
			c.State = ClientDisconnected
			if c.OnStateChange != nil {
				c.OnStateChange(c, ClientDisconnected)
			}
			c.Conn.Close()
			return
		}
		if c.OnMsgRecv != nil {
			c.OnMsgRecv(c, msg, err)
		}
	}
}

func (c *Client) Stop() {
	c.Conn.Close()
	c.State = ClientDisconnected
}

func (c *Client) Send(msg []byte) error {
	return c.encoder.Encode(msg)
}
