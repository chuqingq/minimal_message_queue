package tcp

type ClientState int

const (
	ClientConnected ClientState = iota
	ClientDisconnected
)

type OnClientStateChange func(c *Client, newstate ClientState)
type OnClientMsgRecv func(c *Client, msg []byte, err error)
