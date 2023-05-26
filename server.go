package mmq

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"sync"
	"time"
)

// mmqServer 服务端
type mmqServer struct {
	listener   net.Listener
	outChan    chan Message
	peersMutex sync.Mutex // add/del/dispatch
	peers      []*mmqClient
}

// NewServer 创建并启动服务端
func NewServer(addr string, key, cert, ca []byte) (*mmqServer, error) {
	logger.Debugf("StartServer(%v)", addr)
	server, err := listen(addr, key, cert, ca)
	if err != nil {
		logger.Debugf("Listen error: %v", err)
		server.onStartServer(false, fmt.Sprintf("Listen error: %v", err))
		return nil, err
	}
	logger.Debugf("server[%p].StartServer() success", server)
	server.onStartServer(true, "")
	return server, nil
}

// onStartServer
func (s *mmqServer) onStartServer(success bool, msg string) {
	logger.Debugf("server[%p].onStartServer(%v,%v)", s, success, msg)
	m := NewMessage()
	m.Set("cmd", "onStartServer")
	m.Set("success", success)
	m.Set("msg", msg)
	sendMessage(s.outChan, m)
}

func (s *mmqServer) IsServerAlive() bool {
	return s.listener != nil
}

// listen 服务端启动监听
func listen(addr string, key, cert, ca []byte) (*mmqServer, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logger.Debugf("ResolveTCPAddr(%v) error: %v", addr, err)
		return nil, err
	}
	// pool
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(ca)
	// serverCert
	serverCert, err := tls.X509KeyPair(cert, key)
	if err != nil {
		logger.Debugf("X509KeyPair error: %v", err)
		return nil, err
	}
	// tlsConfig
	tlsConfig := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		ClientCAs:    pool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{serverCert},
	}
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logger.Debugf("ListenTCP() error: %v", err)
		return nil, err
	}
	logger.Debugf("listen success")
	// server
	server := &mmqServer{
		listener: listener,
		outChan:  make(chan Message, 128),
	}
	// accept
	go func() {
		for {
			tcpConn, err := listener.Accept()
			if err != nil {
				logger.Debugf("accept error: %v", err)
				return
			}
			// keepalive
			setKeepAlive(tcpConn, 10*time.Second)
			// setKeepAlive(tcpConn.(*net.TCPConn), 10, 3, 10)
			conn := tls.Server(tcpConn, tlsConfig)
			// handshake
			err = conn.Handshake()
			if err != nil {
				logger.Debugf("server handshake error: %v", err)
				continue
			}
			peer := &mmqClient{
				conn:   conn,
				server: server,
				topics: &mmqTopics{
					topics: make(map[string]interface{}, 16),
				},
			}
			server.addPeer(peer)
			logger.Debugf("server[%p] accepted peer[%p]", server, peer)
			peer.start()
		}
	}()
	return server, nil
}

// Close 关闭服务端
func (s *mmqServer) Close() {
	s.listener.Close()
	s.listener = nil
	for _, c := range s.peers {
		c.Close()
	}
}

func (s *mmqServer) addPeer(c *mmqClient) {
	s.peersMutex.Lock()
	s.peers = append(s.peers, c)
	s.peersMutex.Unlock()
}

func (s *mmqServer) delPeer(c *mmqClient) {
	s.peersMutex.Lock()
	defer s.peersMutex.Unlock()
	for index, peer := range s.peers {
		if c == peer {
			s.peers = append(s.peers[:index], s.peers[index:]...)
			break
		}
	}
}

// dispatchTopic 根据收到的topic分发给订阅的client
func (s *mmqServer) dispatchTopic(topic string, m *Message) {
	logger.Debugf("server[%p].dispatchTopic(%v)", s, topic)
	s.peersMutex.Lock()
	defer s.peersMutex.Unlock()
	for _, peer := range s.peers {
		if peer.topics.match(topic) {
			logger.Debugf("dispatchTopic(%v, %v) to peer: %p", topic, m, peer)
			peer.send(m)
		}
	}
}

func (s *mmqServer) Recv() *Message {
	return recvWithTimeout(s.outChan, -1)
}

func (s *mmqServer) TryRecv() *Message {
	return recvWithTimeout(s.outChan, 0)
}
