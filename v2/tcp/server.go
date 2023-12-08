package tcp

import (
	"encoding/gob"
	"log"
	"net"
	"time"
)

type Server struct {
	ServerAddr string

	OnClientStateChange OnClientStateChange
	OnClientMsgRecv     OnClientMsgRecv

	Listener net.Listener
	Conns    []net.Conn
}

func NewServer(addr string) *Server {
	return &Server{ServerAddr: addr}
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
			log.Printf("Server.Accept() error: %v", err)
			return
		}
		// set keepalive
		conn.(*net.TCPConn).SetKeepAlive(true)
		conn.(*net.TCPConn).SetKeepAlivePeriod(5 * time.Second)

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
		c.encoder = gob.NewEncoder(c.Conn)
		c.decoder = gob.NewDecoder(c.Conn)
		go c.loop()
	}
}

func (s *Server) Stop() {
	s.Listener.Close()
	for _, conn := range s.Conns {
		conn.Close()
	}
}
