package tcpjson

import (
	"encoding/json"
	"net"
)

type Server struct {
	ServerAddr string
	Key        []byte
	Cert       []byte
	Ca         []byte

	OnClientStateChange OnClientStateChange
	OnClientMsgRecv     OnClientMsgRecv

	Listener net.Listener
	Conns    []net.Conn
}

func NewServer(addr string) *Server {
	return &Server{ServerAddr: addr}
}

func (s *Server) SetTLS(key, cert, ca []byte) *Server {
	s.Key = key
	s.Cert = cert
	s.Ca = ca
	return s
}

func (s *Server) SetOnPeerStateChange(handler OnClientStateChange) *Server {
	s.OnClientStateChange = handler
	return s
}

func (s *Server) SetOnMsgRecv(handler OnClientMsgRecv) *Server {
	s.OnClientMsgRecv = handler
	return s
}

func (s *Server) Start() error {
	var err error
	s.Listener, err = net.Listen("tcp", s.ServerAddr)
	if err != nil {
		return err
	}

	go s.loopAccept()
	return nil
}

func (s *Server) loopAccept() {
	s.Conns = make([]net.Conn, 0)
	for {
		conn, err := s.Listener.Accept()
		if err != nil {
			// log.Printf("Server.Accept() error: %v", err)
			return
		}
		s.Conns = append(s.Conns, conn)
		// start peer
		c := &Client{
			Conn:          conn,
			State:         ClientConnected,
			OnStateChange: s.OnClientStateChange,
			OnMsgRecv:     s.OnClientMsgRecv,
		}
		if s.OnClientStateChange != nil {
			s.OnClientStateChange(c, ClientConnected)
		}
		c.encoder = json.NewEncoder(c.Conn)
		c.decoder = json.NewDecoder(c.Conn)
		go c.loop()
	}
}

func (s *Server) Stop() {
	s.Listener.Close()
	for _, conn := range s.Conns {
		conn.Close()
	}
}
