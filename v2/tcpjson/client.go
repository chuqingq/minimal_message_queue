package tcpjson

import (
	"encoding/json"
	"net"
)

type Client struct {
	ServerAddr string
	Key        []byte
	Cert       []byte
	Ca         []byte

	OnStateChange OnClientStateChange
	OnMsgRecv     OnClientMsgRecv

	Conn    net.Conn
	encoder *json.Encoder
	decoder *json.Decoder

	State ClientState
}

func NewClient(serveraddr string) *Client {
	return &Client{ServerAddr: serveraddr, State: ClientDisconnected}
}

func (c *Client) SetTLS(key, cert, ca []byte) *Client {
	c.Key = key
	c.Cert = cert
	c.Ca = ca
	return c
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

	c.encoder = json.NewEncoder(c.Conn)
	c.decoder = json.NewDecoder(c.Conn)
	go c.loop()
	return nil
}

func (c *Client) loop() {
	for {
		var msg json.RawMessage
		err := c.decoder.Decode(&msg)
		if err != nil {
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
	// TODO 对于[]byte直接发还是什么？
	// return c.encoder.Encode(msg)
	_, err := c.Conn.Write(msg)
	return err
}
